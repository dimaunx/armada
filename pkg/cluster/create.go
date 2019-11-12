package cluster

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/dimaunx/armada/pkg/config"
	"github.com/dimaunx/armada/pkg/deploy"
	"github.com/dimaunx/armada/pkg/util"
	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

	kindConfigFilePath, err := util.GenerateKindConfig(cl, configDir, box)
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
			return cl, errors.Errorf("%q: Could not extract version from %s, split is by ':v', example of correct image name: kindest/node:v1.15.3.", cl.Name, flags.ImageName)
		}
	}
	return cl, nil
}

// FinalizeSetup creates custom environment
func FinalizeSetup(cl *config.Cluster, flags *config.Flagpole, box *packr.Box, wg *sync.WaitGroup) error {
	err := util.PrepareKubeConfig(cl)
	if err != nil {
		return err
	}

	kubeConfigFilepath, err := util.GetKubeConfigPath(cl)
	if err != nil {
		return err
	}

	w := wow.New(os.Stdout, spin.Get(spin.Earth), "Finalizing the clusters setup ...")
	w.Start()
	switch cl.Cni {
	case "calico":
		err = deploy.Calico(cl, kubeConfigFilepath, box)
		if err != nil {
			return err
		}
	case "flannel":
		err = deploy.Flannel(cl, "", box)
		if err != nil {
			return err
		}
	case "weave":
		err = deploy.Weave(cl, kubeConfigFilepath, box)
		if err != nil {
			return err
		}
	}
	w.PersistWith(spin.Spinner{Frames: []string{" ✔"}}, fmt.Sprintf(" Cluster %q is ready 🔥🔥🔥", cl.Name))

	if flags.Tiller {
		w := wow.New(os.Stdout, spin.Get(spin.Earth), "Finalizing the clusters setup ...")
		w.Start()
		err = deploy.Tiller(cl, kubeConfigFilepath, box)
		if err != nil {
			return err
		}
		w.PersistWith(spin.Spinner{Frames: []string{" ✔"}}, fmt.Sprintf(" Tiller deployed to %q 📦📦📦", cl.Name))
	}
	wg.Done()
	return nil
}
