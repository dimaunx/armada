package cluster

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/dimaunx/armada/pkg/utils"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/dimaunx/armada/pkg/deploy"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	apiextclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kind "sigs.k8s.io/kind/pkg/cluster"
	kinderrors "sigs.k8s.io/kind/pkg/errors"
)

// Create creates cluster with kind
func Create(cl *config.Cluster, provider *kind.Provider, box *packr.Box, wg *sync.WaitGroup) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	configDir := filepath.Join(currentDir, config.KindConfigDir)
	err = os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		return err
	}

	kindConfigFilePath, err := utils.GenerateKindConfig(cl, configDir, box)
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadFile(kindConfigFilePath)
	if err != nil {
		return err
	}

	log.Infof("Creating cluster %q, cni: %s, podcidr: %s, servicecidr: %s, workers: %v.", cl.Name, cl.Cni, cl.PodSubnet, cl.ServiceSubnet, cl.NumWorkers)

	if err = provider.Create(
		cl.Name,
		kind.CreateWithRawConfig(raw),
		kind.CreateWithNodeImage(cl.NodeImageName),
		kind.CreateWithKubeconfigPath(cl.KubeConfigFilePath),
		kind.CreateWithRetain(cl.Retain),
		kind.CreateWithWaitForReady(cl.WaitForReady),
		kind.CreateWithDisplayUsage(false),
		kind.CreateWithDisplaySalutation(false),
	); err != nil {
		if errs := kinderrors.Errors(err); errs != nil {
			for _, problem := range errs {
				return problem
			}
			return errors.New("aborting due to invalid configuration")
		}
		return errors.Wrap(err, "failed to create cluster")
	}
	wg.Done()
	return nil
}

// FinalizeSetup creates custom environment
func FinalizeSetup(cl *config.Cluster, box *packr.Box, wg *sync.WaitGroup) error {
	masterIP, err := utils.GetMasterDockerIP(cl.Name)
	if err != nil {
		return err
	}

	err = utils.PrepareKubeConfigs(cl.Name, cl.KubeConfigFilePath, masterIP)
	if err != nil {
		return err
	}

	newKubeConfigFilePath, err := utils.GetKubeConfigPath(cl.Name)
	if err != nil {
		return err
	}

	kconfig, err := clientcmd.BuildConfigFromFlags("", newKubeConfigFilePath)
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

	switch cl.Cni {
	case "calico":
		calicoDeploymentFile, err := utils.GenerateCalicoDeploymentFile(cl, box)
		if err != nil {
			return err
		}

		calicoCrdFile, err := box.Resolve("tpl/calico-crd.yaml")
		if err != nil {
			return err
		}

		err = deploy.CrdResources(cl.Name, apiExtClientSet, calicoCrdFile.String())
		if err != nil {
			return err
		}

		err = deploy.Resources(cl.Name, clientSet, calicoDeploymentFile, "Calico")
		if err != nil {
			return err
		}

		err = utils.WaitForDaemonSetReady(cl.Name, clientSet, "kube-system", "calico-node")
		if err != nil {
			return err
		}

		err = utils.WaitForDeploymentReady(cl.Name, clientSet, "kube-system", "coredns")
		if err != nil {
			return err
		}
	case "flannel":
		flannelDeploymentFile, err := utils.GenerateFlannelDeploymentFile(cl, box)
		if err != nil {
			return err
		}

		err = deploy.Resources(cl.Name, clientSet, flannelDeploymentFile, "Flannel")
		if err != nil {
			return err
		}

		err = utils.WaitForDaemonSetReady(cl.Name, clientSet, "kube-system", "kube-flannel-ds-amd64")
		if err != nil {
			return err
		}

		err = utils.WaitForDeploymentReady(cl.Name, clientSet, "kube-system", "coredns")
		if err != nil {
			return err
		}
	case "weave":
		weaveDeploymentFile, err := utils.GenerateWeaveDeploymentFile(cl, box)
		if err != nil {
			return err
		}

		err = deploy.Resources(cl.Name, clientSet, weaveDeploymentFile, "Weave")
		if err != nil {
			return err
		}

		err = utils.WaitForDaemonSetReady(cl.Name, clientSet, "kube-system", "weave-net")
		if err != nil {
			return err
		}

		err = utils.WaitForDeploymentReady(cl.Name, clientSet, "kube-system", "coredns")
		if err != nil {
			return err
		}
	}

	if cl.Tiller {
		err = deploy.Tiller(cl.Name, clientSet, box)
		if err != nil {
			return err
		}
	}
	log.Infof("✔ Cluster %q is ready 🔥🔥🔥", cl.Name)
	wg.Done()
	return nil
}
