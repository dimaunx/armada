package armada

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gobuffalo/packr/v2"

	"github.com/Masterminds/semver"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta "k8s.io/api/extensions/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1beta "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/create"
	"sigs.k8s.io/kind/pkg/util"
)

const KindConfigDir = "output/kind-clusters"
const LocalKubeConfigDir = "output/kind-config/local-dev"
const ContainerKubeConfigDir = "output/kind-config/container"

type flagpole struct {
	ImageName   string
	Wait        time.Duration
	Retain      bool
	Weave       bool
	Flannel     bool
	Calico      bool
	Debug       bool
	Tiller      bool
	NumClusters int
}

type Cluster struct {
	Cni                 string
	Name                string
	PodSubnet           string
	ServiceSubnet       string
	DNSDomain           string
	KubeAdminApiVersion string
}

type KubeConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Cluster struct {
			CertificateAuthorityData string `yaml:"certificate-authority-data"`
			Server                   string `yaml:"server"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	} `yaml:"clusters"`
	Contexts []struct {
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
	Kind           string `yaml:"kind"`
	Preferences    struct {
	} `yaml:"preferences"`
	Users []struct {
		Name string `yaml:"name"`
		User struct {
			ClientCertificateData string `yaml:"client-certificate-data"`
			ClientKeyData         string `yaml:"client-key-data"`
		} `yaml:"user"`
	} `yaml:"users"`
}

// Get different kubeconfig paths for local and docker based runs
func (cl *Cluster) GetKubeConfigPath() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Dummy destination
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ctx := context.Background()
	dockerCli, err := dockerclient.NewEnvClient()
	if err != nil {
		return "", err
	}

	networkFilter := filters.NewArgs()
	networkFilter.Add("driver", "bridge")
	networks, err := dockerCli.NetworkList(ctx, types.NetworkListOptions{Filters: networkFilter})
	if err != nil {
		return "", err
	}

	for _, network := range networks {
		dockerNet := network.IPAM.Config[0].Subnet
		_, ipNet, err := net.ParseCIDR(dockerNet)
		if err != nil {
			return "", err
		}

		if ipNet.Contains(localAddr.IP) {
			log.Debugf("Running in a container. Bridge network: %s, ip: %s.	", dockerNet, localAddr.IP)
			kubeConfigFilePath := filepath.Join(currentDir, ContainerKubeConfigDir, strings.Join([]string{"kind-config", cl.Name}, "-"))
			return kubeConfigFilePath, nil
		}
	}
	log.Debugf("Running in a host mode. ip: %s.", localAddr.IP)
	kubeConfigFilePath := filepath.Join(currentDir, LocalKubeConfigDir, strings.Join([]string{"kind-config", cl.Name}, "-"))
	return kubeConfigFilePath, nil
}

// Get control plain master docker internal ip
func (cl *Cluster) GetMasterDockerIp() (string, error) {
	ctx := context.Background()
	dockerCli, err := dockerclient.NewEnvClient()
	if err != nil {
		return "", err
	}

	containerFilter := filters.NewArgs()
	containerFilter.Add("name", strings.Join([]string{cl.Name, "control-plane"}, "-"))
	containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{
		Filters: containerFilter,
		Limit:   1,
	})
	if err != nil {
		return "", err
	}
	return containers[0].NetworkSettings.Networks["bridge"].IPAddress, nil
}

//Create kind cluster
func (cl *Cluster) CreateCluster(flags *flagpole, cf string) error {
	// create a cluster context and create the cluster
	ctx := cluster.NewContext(cl.Name)
	log.Infof("Creating cluster %q, cni: %s, podcidr: %s, servicecidr: %s ...", cl.Name, cl.Cni, cl.PodSubnet, cl.ServiceSubnet)

	if err := ctx.Create(
		create.WithConfigFile(cf),
		create.WithNodeImage(flags.ImageName),
		create.Retain(flags.Retain),
		create.WaitForReady(flags.Wait),
	); err != nil {
		if utilErrors, ok := err.(util.Errors); ok {
			for _, problem := range utilErrors.Errors() {
				return problem
			}
			return errors.New("aborting due to invalid configuration")
		}
		return errors.Wrapf(err, "failed to create cluster %q", cl.Name)
	}
	return nil
}

//Generate config file
func (cl *Cluster) GenerateConfig(cf string, ct string) error {
	t, err := template.New("config").Parse(ct)
	if err != nil {
		return err
	}

	f, err := os.Create(cf)
	if err != nil {
		return errors.Wrapf(err, "creating config file %s", cf)
	}

	err = t.Execute(f, cl)
	if err != nil {
		return errors.Wrapf(err, "creating config file %s", cf)
	}

	if err := f.Close(); err != nil {
		return err
	}

	log.Debugf("ClustersConfig files for %s generated.", cl.Name)
	return nil
}

// Modify kube config
func (cl *Cluster) PrepareKubeConfig() error {
	var kubeconf KubeConfig

	currentDir, _ := os.Getwd()
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "failed to get current user")
	}

	_ = os.MkdirAll(LocalKubeConfigDir, os.ModePerm)
	_ = os.MkdirAll(ContainerKubeConfigDir, os.ModePerm)
	kindKubeFileName := strings.Join([]string{"kind-config", cl.Name}, "-")
	kindKubeFilePath := filepath.Join(usr.HomeDir, ".kube", kindKubeFileName)
	newLocalKubeFile := filepath.Join(currentDir, LocalKubeConfigDir, kindKubeFileName)
	newContainerKubeFile := filepath.Join(currentDir, ContainerKubeConfigDir, kindKubeFileName)

	kubeFile, err := ioutil.ReadFile(kindKubeFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read kube config %s.", kindKubeFilePath)
	}

	err = yaml.Unmarshal(kubeFile, &kubeconf)
	if err != nil {
		return errors.Wrapf(err, "failed to read kube config %s.", kindKubeFilePath)
	}

	kubeconf.CurrentContext = cl.Name
	kubeconf.Contexts[0].Name = cl.Name
	kubeconf.Contexts[0].Context.Cluster = cl.Name
	kubeconf.Contexts[0].Context.User = cl.Name
	kubeconf.Clusters[0].Name = cl.Name
	kubeconf.Users[0].Name = cl.Name

	d, err := yaml.Marshal(&kubeconf)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal kube config.")
	}

	err = ioutil.WriteFile(newLocalKubeFile, d, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to save kube config %s.", newLocalKubeFile)
	}

	masterIp, err := cl.GetMasterDockerIp()
	if err != nil {
		return err
	}

	kubeconf.Clusters[0].Cluster.Server = "https://" + masterIp + ":6443"
	d, err = yaml.Marshal(&kubeconf)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal kube config.")
	}

	err = ioutil.WriteFile(newContainerKubeFile, d, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to save kube config %s.", newContainerKubeFile)
	}
	return nil
}

// Wait for coredns deployment
func (cl *Cluster) WaitForCoreDnsDeployment(kf string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kf)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	ctx := context.Background()
	weaveTimeout := 5 * time.Minute
	log.Infof("Waiting up to %v for coredns pods to be ready %s...", weaveTimeout, cl.Name)
	corednsContext, cancel := context.WithTimeout(ctx, weaveTimeout)
	wait.Until(func() {
		corednsDeployment, err := clientset.AppsV1().Deployments("kube-system").Get("coredns", metav1.GetOptions{})
		if err == nil && corednsDeployment.Status.ReadyReplicas > 0 {
			if corednsDeployment.Status.ReadyReplicas == 2 {
				log.Infof("✔ Coredns successfully deployed to %s, ready replicas: %v", cl.Name, corednsDeployment.Status.ReadyReplicas)
				cancel()
			} else {
				log.Warnf("Still waiting for coredns deployment %s, ready replicas: %v", cl.Name, corednsDeployment.Status.ReadyReplicas)
			}
		} else {
			log.Warnf("Still waiting for coredns deployment %s.", cl.Name)
		}
	}, 10*time.Second, corednsContext.Done())
	err = corednsContext.Err()
	if err != nil && err != context.Canceled {
		return errors.Wrap(err, "Error waiting for coredns deployment.")
	}
	return nil
}

// Create tiller deployment
func (cl *Cluster) WaitForTillerDeployment(kf string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kf)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	ctx := context.Background()
	tillerTimeout := 5 * time.Minute
	log.Infof("Waiting up to %v for tiller pods to be ready %s...", tillerTimeout, cl.Name)
	tillerContext, cancel := context.WithTimeout(ctx, tillerTimeout)
	wait.Until(func() {
		tillerDeployment, err := clientset.ExtensionsV1beta1().Deployments("kube-system").Get("tiller-deploy", metav1.GetOptions{})
		if err == nil && tillerDeployment.Status.ReadyReplicas > 0 {
			if tillerDeployment.Status.ReadyReplicas == 1 {
				log.Infof("✔ Tiller successfully deployed to %s, ready replicas: %v", cl.Name, tillerDeployment.Status.ReadyReplicas)
				cancel()
			} else {
				log.Warnf("Still waiting for tiller deployment %s, ready replicas: %v", cl.Name, tillerDeployment.Status.ReadyReplicas)
			}
		} else {
			log.Warnf("Still waiting for tiller deployment for %s.", cl.Name)
		}
	}, 10*time.Second, tillerContext.Done())

	err = tillerContext.Err()
	if err != nil && err != context.Canceled {
		return errors.Wrap(err, "Error waiting for tiller deployment.")
	}
	return nil
}

// Deploy resources from yaml file
func (cl *Cluster) DeployResources(df string, kf string, rn string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kf)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	acceptedK8sTypes := regexp.MustCompile(`(Role|RoleBinding|ClusterRole|ClusterRoleBinding|ServiceAccount|ConfigMap|DaemonSet|Deployment|Service|Pod)`)
	fileAsString := df[:]
	sepYamlfiles := strings.Split(fileAsString, "---")
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(f), nil, nil)

		if err != nil {
			return errors.Wrap(err, "Error while decoding YAML object. Err was: ")
		}

		if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
			log.Warnf("The file contains K8s object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind)
		} else {
			switch o := obj.(type) {
			case *corev1.ServiceAccount:
				result, err := clientSet.CoreV1().ServiceAccounts(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ ServiceAccount %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *rbacv1.Role:
				result, err := clientSet.RbacV1().Roles(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ Role %s created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *rbacv1.RoleBinding:
				result, err := clientSet.RbacV1().RoleBindings(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ RoleBinding %s created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *rbacv1.ClusterRole:
				result, err := clientSet.RbacV1().ClusterRoles().Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ ClusterRole %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *rbacv1.ClusterRoleBinding:
				result, err := clientSet.RbacV1().ClusterRoleBindings().Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ ClusterRoleBinding %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *corev1.ConfigMap:
				result, err := clientSet.CoreV1().ConfigMaps(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ ConfigMap %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *corev1.Service:
				result, err := clientSet.CoreV1().Services(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ Service %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *corev1.Pod:
				result, err := clientSet.CoreV1().Pods(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ Pod %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *policyv1beta1.PodSecurityPolicy:
				result, err := clientSet.PolicyV1beta1().PodSecurityPolicies().Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ PodSecurityPolicy %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *appsv1.DaemonSet:
				result, err := clientSet.AppsV1().DaemonSets(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ Daemonset %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *extv1beta.DaemonSet:
				result, err := clientSet.ExtensionsV1beta1().DaemonSets(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ Daemonset %s was created for %s at: %s.", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *appsv1.Deployment:
				result, err := clientSet.AppsV1().Deployments(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ Deployment %s was created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			case *extv1beta.Deployment:
				result, err := clientSet.ExtensionsV1beta1().Deployments(o.Namespace).Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ Deployment %s created for %s at: %s", o.Name, cl.Name, result.CreationTimestamp)
				}
			}
		}
	}
	log.Infof("✔ %s was deployed to %s.", rn, cl.Name)
	return nil
}

// Deploy CRD resources from yaml file
func (cl *Cluster) DeployCrdResources(cf string, kf string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kf)
	if err != nil {
		return err
	}

	apiExtClientSet, err := apiextclientset.NewForConfig(config)
	if err != nil {
		return err
	}

	acceptedK8sTypes := regexp.MustCompile(`(CustomResourceDefinition)`)
	fileAsString := cf[:]
	sepYamlfiles := strings.Split(fileAsString, "---")
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		decode := apiextscheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(f), nil, nil)

		if err != nil {
			return errors.Wrap(err, "Error while decoding YAML object. Err was: ")
		}

		if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
			log.Warnf("The file contains K8s object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind)
		} else {
			switch o := obj.(type) {
			case *apiextv1beta.CustomResourceDefinition:
				_, err := apiExtClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Create(o)
				if err != nil && !apierr.IsAlreadyExists(err) {
					return err
				} else {
					log.Debugf("✔ CRD %s was created for %s.", o.Name, cl.Name)
				}
			}
		}
	}
	return nil
}

func CreateEnvironment(i int, flags *flagpole, wg *sync.WaitGroup) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	podIp := net.ParseIP("10.0.0.0")
	podIp = podIp.To4()
	podIp[1] += byte(4 * i)

	serviceIp := net.ParseIP("100.0.0.0")
	serviceIp = serviceIp.To4()
	serviceIp[1] += byte(i)

	cl := &Cluster{
		Cni:                 "kindnet",
		Name:                "cl" + strconv.Itoa(i),
		PodSubnet:           podIp.String() + "/14",
		ServiceSubnet:       serviceIp.String() + "/16",
		DNSDomain:           "cl" + strconv.Itoa(i) + ".local",
		KubeAdminApiVersion: "kubeadm.k8s.io/v1beta2",
	}

	if flags.ImageName != "" {
		tgt := semver.MustParse("1.15")
		results := strings.Split(flags.ImageName, ":v")
		if len(results) >= 2 {
			sver := semver.MustParse(results[len(results)-1])
			if sver.LessThan(tgt) {
				cl.KubeAdminApiVersion = "kubeadm.k8s.io/v1beta1"
			}
		} else {
			return errors.Errorf("%q: Could not extract version from %s, split is by ':v', example of correct image name: kindest/node:v1.14.6.", cl.Name, flags.ImageName)
		}
	}

	configDir := filepath.Join(currentDir, "output/kind-clusters")
	err = os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "%s", cl.Name)
	}

	box := packr.New("manifests", "../manifests")

	clusterConfigTemplate, err := box.Resolve("tpl/cluster-config.yaml")
	if err != nil {
		return errors.Wrapf(err, "%s", cl.Name)
	}

	kindConfigFilePath := filepath.Join(configDir, cl.Name+"-kind-config.yaml")
	kubeConfigFilePath, err := cl.GetKubeConfigPath()
	if err != nil {
		return errors.Wrapf(err, "%s", cl.Name)
	}

	if flags.Weave {
		cl.Cni = "weave"
		flags.Wait = 0

		err = cl.GenerateConfig(kindConfigFilePath, clusterConfigTemplate.String())
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.CreateCluster(flags, kindConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.PrepareKubeConfig()
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		weaveDeploymentTemplate, err := box.Resolve("tpl/weave-daemonset.yaml")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		t, err := template.New("weave").Parse(weaveDeploymentTemplate.String())
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		var weaveDeploymentFile bytes.Buffer
		err = t.Execute(&weaveDeploymentFile, cl)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.DeployResources(weaveDeploymentFile.String(), kubeConfigFilePath, "Weave")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.WaitForCoreDnsDeployment(kubeConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

	} else if flags.Flannel {
		cl.Cni = "flannel"
		flags.Wait = 0

		err = cl.GenerateConfig(kindConfigFilePath, clusterConfigTemplate.String())
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.CreateCluster(flags, kindConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.PrepareKubeConfig()
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		flannelDeploymentTemplate, err := box.Resolve("tpl/flannel-daemonset.yaml")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		t, err := template.New("flannel").Parse(flannelDeploymentTemplate.String())
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		var flannelDeploymentFile bytes.Buffer
		err = t.Execute(&flannelDeploymentFile, cl)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.DeployResources(flannelDeploymentFile.String(), kubeConfigFilePath, "Flannel")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.WaitForCoreDnsDeployment(kubeConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

	} else if flags.Calico {
		cl.Cni = "calico"
		flags.Wait = 0

		err = cl.GenerateConfig(kindConfigFilePath, clusterConfigTemplate.String())
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.CreateCluster(flags, kindConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.PrepareKubeConfig()
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		calicoCrdFile, err := box.Resolve("tpl/calico-crd.yaml")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.DeployCrdResources(calicoCrdFile.String(), kubeConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		calicoDeploymentTemplate, err := box.Resolve("tpl/calico-daemonset.yaml")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		t, err := template.New("calico").Parse(calicoDeploymentTemplate.String())
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		var calicoDeploymentFile bytes.Buffer
		err = t.Execute(&calicoDeploymentFile, cl)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.DeployResources(calicoDeploymentFile.String(), kubeConfigFilePath, "Calico")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.WaitForCoreDnsDeployment(kubeConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

	} else {
		err = cl.GenerateConfig(kindConfigFilePath, clusterConfigTemplate.String())
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.CreateCluster(flags, kindConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.PrepareKubeConfig()
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}
	}

	if flags.Tiller {
		tillerDeploymentFile, err := box.Resolve("helm/tiller-deployment.yaml")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.DeployResources(tillerDeploymentFile.String(), kubeConfigFilePath, "Tiller")
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}

		err = cl.WaitForTillerDeployment(kubeConfigFilePath)
		if err != nil {
			return errors.Wrapf(err, "%s", cl.Name)
		}
	}
	wg.Done()
	return nil
}

func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "create",
		Short: "Creates e2e environment",
		Long:  "Creates multiple kind clusters",
	}

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	cmd.AddCommand(CreateClustersCommand())
	return cmd
}

func CreateClustersCommand() *cobra.Command {
	flags := &flagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "clusters",
		Short: "Creates multiple kubernetes clusters",
		Long:  "Creates multiple kubernetes clusters using Docker container 'nodes'",
		RunE: func(cmd *cobra.Command, args []string) error {

			if flags.Debug {
				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)
			}

			var wg sync.WaitGroup
			wg.Add(flags.NumClusters)
			for i := 1; i <= flags.NumClusters; i++ {
				go func(i int) {
					known, err := cluster.IsKnown("cl" + strconv.Itoa(i))
					if err != nil {
						log.Fatal(err)
					}
					if known {
						log.Infof("✔ Cluster with the name %q already exists", "cl"+strconv.Itoa(i))
						wg.Done()
					} else {
						err := CreateEnvironment(i, flags, &wg)
						if err != nil {
							defer wg.Done()
							log.Error(err)
						}
					}
				}(i)
			}
			wg.Wait()
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			files, err := ioutil.ReadDir(KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range files {
				clName := strings.Split(file.Name(), "-")[0]
				known, err := cluster.IsKnown(clName)
				if err != nil {
					log.Error(err)
				}
				if known {
					log.Debugf("✔ Cluster with the name %q already exists", clName)
				} else {
					cl := &Cluster{Name: clName}
					err = cl.PrepareKubeConfig()
					if err != nil {
						log.Errorf("test %s", err)
					}
				}
			}
			log.Infof("✔ Kubeconfigs: export KUBECONFIG=$(echo ./output/kind-config/local-dev/kind-config-cl{1..%v} | sed 's/ /:/g')", flags.NumClusters)
		},
	}
	cmd.Flags().StringVarP(&flags.ImageName, "image", "i", "", "node docker image to use for booting the cluster")
	cmd.Flags().BoolVarP(&flags.Retain, "retain", "", true, "retain nodes for debugging when cluster creation fails")
	cmd.Flags().BoolVarP(&flags.Weave, "weave", "w", false, "deploy weave")
	cmd.Flags().BoolVarP(&flags.Tiller, "tiller", "t", false, "deploy tiller")
	cmd.Flags().BoolVarP(&flags.Calico, "calico", "c", false, "deploy calico")
	cmd.Flags().BoolVarP(&flags.Flannel, "flannel", "f", false, "deploy flannel")
	cmd.Flags().BoolVarP(&flags.Debug, "debug", "v", false, "set log level to debug")
	cmd.Flags().DurationVar(&flags.Wait, "wait", 5*time.Minute, "amount of minutes to wait for control plane nodes to be ready")
	cmd.Flags().IntVarP(&flags.NumClusters, "num", "n", 2, "number of clusters to create")
	return cmd
}
