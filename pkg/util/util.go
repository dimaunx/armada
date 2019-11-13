package util

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/dimaunx/armada/pkg/config"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type kubeConfig struct {
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

// PrepareKubeConfig modifies kubconfig file generated by kind and returns the kubeconfig file path
func PrepareKubeConfig(cl *config.Cluster) error {
	var kubeconf kubeConfig

	currentDir, _ := os.Getwd()
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "failed to get current user")
	}

	_ = os.MkdirAll(config.LocalKubeConfigDir, os.ModePerm)
	_ = os.MkdirAll(config.ContainerKubeConfigDir, os.ModePerm)
	kindKubeFileName := strings.Join([]string{"kind-config", cl.Name}, "-")
	kindKubeFilePath := filepath.Join(usr.HomeDir, ".kube", kindKubeFileName)
	newLocalKubeFilePath := filepath.Join(currentDir, config.LocalKubeConfigDir, kindKubeFileName)
	newContainerKubeFilePath := filepath.Join(currentDir, config.ContainerKubeConfigDir, kindKubeFileName)

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

	err = ioutil.WriteFile(newLocalKubeFilePath, d, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to save kube config %s.", newLocalKubeFilePath)
	}

	masterIP, err := GetMasterDockerIP(cl)
	if err != nil {
		return err
	}

	kubeconf.Clusters[0].Cluster.Server = "https://" + masterIP + ":6443"
	d, err = yaml.Marshal(&kubeconf)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal kube config.")
	}

	err = ioutil.WriteFile(newContainerKubeFilePath, d, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to save kube config %s.", newContainerKubeFilePath)
	}
	return nil
}

// GetKubeConfigPath returns different kubeconfig paths for local and docker based runs
func GetKubeConfigPath(cl *config.Cluster) (string, error) {
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
	networks, err := dockerCli.NetworkList(ctx, dockertypes.NetworkListOptions{Filters: networkFilter})
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
			kubeConfigFilePath := filepath.Join(currentDir, config.ContainerKubeConfigDir, strings.Join([]string{"kind-config", cl.Name}, "-"))
			return kubeConfigFilePath, nil
		}
	}
	log.Debugf("Running in a host mode. ip: %s.", localAddr.IP)
	kubeConfigFilePath := filepath.Join(currentDir, config.LocalKubeConfigDir, strings.Join([]string{"kind-config", cl.Name}, "-"))
	return kubeConfigFilePath, nil
}

// GetMasterDockerIP gets control plain master docker internal ip
func GetMasterDockerIP(cl *config.Cluster) (string, error) {
	ctx := context.Background()
	dockerCli, err := dockerclient.NewEnvClient()
	if err != nil {
		return "", err
	}

	containerFilter := filters.NewArgs()
	containerFilter.Add("name", strings.Join([]string{cl.Name, "control-plane"}, "-"))
	containers, err := dockerCli.ContainerList(ctx, dockertypes.ContainerListOptions{
		Filters: containerFilter,
		Limit:   1,
	})
	if err != nil {
		return "", err
	}
	return containers[0].NetworkSettings.Networks["bridge"].IPAddress, nil
}

// iterate func map for config template
func iterate(start, end int) (stream chan int) {
	stream = make(chan int)
	go func() {
		for i := start; i <= end; i++ {
			stream <- i
		}
		close(stream)
	}()
	return
}

// GenerateKindConfig creates kind config file and returns its path
func GenerateKindConfig(cl *config.Cluster, configDir string, box *packr.Box) (string, error) {
	kindConfigFileTemplate, err := box.Resolve("tpl/cluster-config.yaml")
	if err != nil {
		return "", err
	}

	t, err := template.New("config").Funcs(template.FuncMap{"iterate": iterate}).Parse(kindConfigFileTemplate.String())
	if err != nil {
		return "", err
	}

	kindConfigFilePath := filepath.Join(configDir, "kind-config-"+cl.Name+".yaml")
	f, err := os.Create(kindConfigFilePath)
	if err != nil {
		return "", err
	}

	err = t.Execute(f, cl)
	if err != nil {
		return "", err
	}

	if err := f.Close(); err != nil {
		return "", err
	}
	log.Debugf("ClustersConfig files for %s generated.", cl.Name)
	return kindConfigFilePath, nil
}
