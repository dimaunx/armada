package cluster

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/dimaunx/armada/pkg/wait"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Masterminds/semver"
	"github.com/dimaunx/armada/pkg/config"
	"github.com/dimaunx/armada/pkg/deploy"
	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	apiextclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/create"
	kindutil "sigs.k8s.io/kind/pkg/util"
)

// Create creates cluster with kind
func Create(cl *config.Cluster, flags *config.Flagpole, box *packr.Box, wg *sync.WaitGroup) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	configDir := filepath.Join(currentDir, config.KindConfigDir)
	err = os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		return err
	}

	kindConfigFilePath, err := GenerateKindConfig(cl, configDir, box)
	if err != nil {
		return err
	}

	ctx := cluster.NewContext(cl.Name)
	log.Infof("Creating cluster %q, cni: %s, podcidr: %s, servicecidr: %s, workers: %v.", cl.Name, cl.Cni, cl.PodSubnet, cl.ServiceSubnet, cl.NumWorkers)

	if err := ctx.Create(
		create.WithConfigFile(kindConfigFilePath),
		create.WithNodeImage(flags.ImageName),
		create.Retain(flags.Retain),
		create.WaitForReady(flags.Wait),
	); err != nil {
		if utilErrors, ok := err.(kindutil.Errors); ok {
			for _, problem := range utilErrors.Errors() {
				return problem
			}
			return errors.New("aborting due to invalid configuration")
		}
		return errors.Wrapf(err, "failed to create cluster %q", cl.Name)
	}
	wg.Done()
	return nil
}

// PopulateClusterConfig return a desired cluster object
func PopulateClusterConfig(i int, flags *config.Flagpole) (*config.Cluster, error) {

	cl := &config.Cluster{
		Name:                config.ClusterNameBase + strconv.Itoa(i),
		NumWorkers:          config.NumWorkers,
		DNSDomain:           config.ClusterNameBase + strconv.Itoa(i) + ".local",
		KubeAdminAPIVersion: config.KubeAdminAPIVersion,
	}

	podIP := net.ParseIP(config.PodCidrBase)
	podIP = podIP.To4()
	serviceIP := net.ParseIP(config.ServiceCidrBase)
	serviceIP = serviceIP.To4()

	if !flags.Overlap {
		podIP[1] += byte(4 * i)
		serviceIP[1] += byte(i)
	}

	cl.PodSubnet = podIP.String() + config.PodCidrMask
	cl.ServiceSubnet = serviceIP.String() + config.ServiceCidrMask

	if flags.Weave {
		cl.Cni = "weave"
		flags.Wait = 0
	} else if flags.Calico {
		cl.Cni = "calico"
		flags.Wait = 0
	} else if flags.Flannel {
		cl.Cni = "flannel"
		flags.Wait = 0
	} else {
		cl.Cni = "kindnet"
	}

	if flags.ImageName != "" {
		tgt := semver.MustParse("1.15")
		results := strings.Split(flags.ImageName, ":v")
		if len(results) == 2 {
			sver := semver.MustParse(results[len(results)-1])
			if sver.LessThan(tgt) {
				cl.KubeAdminAPIVersion = "kubeadm.k8s.io/v1beta1"
			}
		} else {
			return nil, errors.Errorf("%q: Could not extract version from %s, split is by ':v', example of correct image name: kindest/node:v1.15.3.", cl.Name, flags.ImageName)
		}
	}
	return cl, nil
}

// FinalizeSetup creates custom environment
func FinalizeSetup(cl *config.Cluster, flags *config.Flagpole, box *packr.Box, wg *sync.WaitGroup) error {
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "failed to get current user")
	}

	kindKubeFileName := strings.Join([]string{"kind-config", cl.Name}, "-")
	kindKubeFilePath := filepath.Join(usr.HomeDir, ".kube", kindKubeFileName)

	masterIP, err := GetMasterDockerIP(cl.Name)
	if err != nil {
		return err
	}

	err = PrepareKubeConfig(cl.Name, kindKubeFilePath, masterIP)
	if err != nil {
		return err
	}

	kubeConfigFilePath, err := GetKubeConfigPath(cl.Name)
	if err != nil {
		return err
	}

	kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		return err
	}

	apiExtClientSet, err := apiextclientset.NewForConfig(kconfig)
	if err != nil {
		return err
	}

	w := wow.New(os.Stdout, spin.Get(spin.Earth), "Finalizing the clusters setup ...")
	w.Start()
	switch cl.Cni {
	case "calico":
		calicoDeploymentFile, err := GenerateCalicoDeploymentFile(cl, box)
		if err != nil {
			return err
		}

		calicoCrdFile, err := box.Resolve("tpl/calico-crd.yaml")
		if err != nil {
			return err
		}

		err = deploy.CreateCrdResources(cl.Name, apiExtClientSet, calicoCrdFile.String())
		if err != nil {
			return err
		}

		err = deploy.CreateResources(cl.Name, clientSet, calicoDeploymentFile, "Calico")
		if err != nil {
			return err
		}

		err = wait.ForDaemonSetReady(cl.Name, clientSet, "kube-system", "calico-node")
		if err != nil {
			return err
		}

		err = wait.ForDeploymentReady(cl.Name, clientSet, "kube-system", "coredns")
		if err != nil {
			return err
		}
	case "flannel":
		flannelDeploymentFile, err := GenerateFlannelDeploymentFile(cl, box)
		if err != nil {
			return err
		}

		err = deploy.CreateResources(cl.Name, clientSet, flannelDeploymentFile, "Flannel")
		if err != nil {
			return err
		}

		err = wait.ForDaemonSetReady(cl.Name, clientSet, "kube-system", "kube-flannel-ds-amd64")
		if err != nil {
			return err
		}

		err = wait.ForDeploymentReady(cl.Name, clientSet, "kube-system", "coredns")
		if err != nil {
			return err
		}
	case "weave":
		weaveDeploymentFile, err := GenerateWeaveDeploymentFile(cl, box)
		if err != nil {
			return err
		}

		err = deploy.CreateResources(cl.Name, clientSet, weaveDeploymentFile, "Weave")
		if err != nil {
			return err
		}

		err = wait.ForDaemonSetReady(cl.Name, clientSet, "kube-system", "weave-net")
		if err != nil {
			return err
		}

		err = wait.ForDeploymentReady(cl.Name, clientSet, "kube-system", "coredns")
		if err != nil {
			return err
		}
	}

	if flags.Tiller {
		err = deploy.Tiller(cl.Name, clientSet, box)
		if err != nil {
			return err
		}
	}
	w.PersistWith(spin.Spinner{Frames: []string{" ✔"}}, fmt.Sprintf(" Cluster %q is ready 🔥🔥🔥", cl.Name))
	wg.Done()
	return nil
}
