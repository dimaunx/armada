package cluster

import (
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/semver"

	"github.com/dimaunx/armada/pkg/defaults"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Config type
type Config struct {
	// Cni is a name of the cni that will be installed for a cluster
	Cni string

	// Name is a cluster name
	Name string

	// PodSubnet is pod subnet cidr and mask
	PodSubnet string

	// ServiceSubnet is a service subnet cidr and mask
	ServiceSubnet string

	// DNSDomain is cluster dns domain name
	DNSDomain string

	// // KubeAdminAPIVersion for each cluster
	KubeAdminAPIVersion string

	// NumWorkers is the number of worker nodes
	NumWorkers int

	// KubeConfigFilePath is the destination where kind will generate the original kubeconfig file
	KubeConfigFilePath string

	// Amount of time to wait for control plain to be ready
	WaitForReady time.Duration

	// Config image name
	NodeImageName string

	// Retain if to retain the cluster despite and error
	Retain bool

	// Tiller if to deploy a cluster with tiller
	Tiller bool
}

// GenerateKindConfig creates kind config file and returns its path
func GenerateKindConfig(cl *Config, configDir string, box *packr.Box) (string, error) {
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
	log.Debugf("Config config file for %s generated.", cl.Name)
	return kindConfigFilePath, nil
}

// PopulateConfig return a desired cluster config object
func PopulateConfig(i int, image, cni string, retain, tiller, overlap bool, wait time.Duration) (*Config, error) {

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	cl := &Config{
		Name:                defaults.ClusterNameBase + strconv.Itoa(i),
		NodeImageName:       image,
		Cni:                 cni,
		NumWorkers:          defaults.NumWorkers,
		DNSDomain:           defaults.ClusterNameBase + strconv.Itoa(i) + ".local",
		KubeAdminAPIVersion: defaults.KubeAdminAPIVersion,
		Retain:              retain,
		Tiller:              tiller,
		WaitForReady:        wait,
		KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", strings.Join([]string{"kind-config", defaults.ClusterNameBase + strconv.Itoa(i)}, "-")),
	}

	podIP := net.ParseIP(defaults.PodCidrBase)
	podIP = podIP.To4()
	serviceIP := net.ParseIP(defaults.ServiceCidrBase)
	serviceIP = serviceIP.To4()

	if !overlap {
		podIP[1] += byte(4 * i)
		serviceIP[1] += byte(i)
	}

	cl.PodSubnet = podIP.String() + defaults.PodCidrMask
	cl.ServiceSubnet = serviceIP.String() + defaults.ServiceCidrMask

	if cni != "kindnet" {
		cl.WaitForReady = 0
	}

	if image != "" {
		tgt := semver.MustParse("1.15")
		results := strings.Split(image, ":v")
		if len(results) == 2 {
			sver := semver.MustParse(results[len(results)-1])
			if sver.LessThan(tgt) {
				cl.KubeAdminAPIVersion = "kubeadm.k8s.io/v1beta1"
			}
		} else {
			return nil, errors.Errorf("%q: Could not extract version from %s, split is by ':v', example of correct image name: kindest/node:v1.15.3.", cl.Name, cl.NodeImageName)
		}
	}
	return cl, nil
}
