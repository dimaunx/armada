package cluster

import (
	"os"
	"path/filepath"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/kind/pkg/cluster"
)

// Destroy destroys a kind cluster
func Destroy(clName string) error {
	log.Infof("Deleting cluster %q ...\n", clName)
	ctx := cluster.NewContext(clName)
	if err := ctx.Delete(); err != nil {
		return errors.Wrapf(err, "failed to delete cluster %s", clName)
	}

	_ = os.Remove(filepath.Join(config.KindConfigDir, "kind-config-"+clName+".yaml"))
	_ = os.Remove(filepath.Join(config.LocalKubeConfigDir, "kind-config-"+clName))
	_ = os.Remove(filepath.Join(config.ContainerKubeConfigDir, "kind-config-"+clName))

	return nil
}
