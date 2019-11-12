package cluster

import (
	"os"
	"path/filepath"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/kind/pkg/cluster"
)

// Delete deletes a kind cluster
func Delete(cl *config.Cluster, clusterConfigFile os.FileInfo) error {
	log.Infof("Deleting cluster %q ...\n", cl.Name)
	ctx := cluster.NewContext(cl.Name)
	if err := ctx.Delete(); err != nil {
		return errors.Wrapf(err, "failed to delete cluster %s", cl.Name)
	}
	_ = os.Remove(filepath.Join(config.KindConfigDir, clusterConfigFile.Name()))
	_ = os.Remove(filepath.Join(config.LocalKubeConfigDir, "kind-config-"+cl.Name))
	_ = os.Remove(filepath.Join(config.ContainerKubeConfigDir, "kind-config-"+cl.Name))

	return nil
}
