package cluster

import (
	"os"
	"path/filepath"
	"time"

	"github.com/dimaunx/armada/pkg/constants"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/create"
	"sigs.k8s.io/kind/pkg/util"
)

// Cluster type
type Cluster struct {
	Cni                 string
	Name                string
	PodSubnet           string
	ServiceSubnet       string
	DNSDomain           string
	KubeAdminAPIVersion string
	NumWorkers          int
}

// Flagpole type
type Flagpole struct {
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

// CreateKindCluster creates cluster with kind
func CreateKindCluster(cl *Cluster, flags Flagpole, clusterConfigFile string) error {
	// create a cluster context and create the cluster
	ctx := cluster.NewContext(cl.Name)
	log.Infof("Creating cluster %q, cni: %s, podcidr: %s, servicecidr: %s, workers: %v.", cl.Name, cl.Cni, cl.PodSubnet, cl.ServiceSubnet, cl.NumWorkers)

	if err := ctx.Create(
		create.WithConfigFile(clusterConfigFile),
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

// DeleteKindCluster delete a kind cluster
func DeleteKindCluster(cl *Cluster, clusterConfigFile os.FileInfo) error {
	log.Infof("Deleting cluster %q ...\n", cl.Name)
	ctx := cluster.NewContext(cl.Name)
	if err := ctx.Delete(); err != nil {
		return errors.Wrapf(err, "failed to delete cluster %s", cl.Name)
	}
	_ = os.Remove(filepath.Join(constants.KindConfigDir, clusterConfigFile.Name()))
	_ = os.Remove(filepath.Join(constants.LocalKubeConfigDir, "kind-config-"+cl.Name))
	_ = os.Remove(filepath.Join(constants.ContainerKubeConfigDir, "kind-config-"+cl.Name))

	return nil
}
