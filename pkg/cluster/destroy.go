package cluster

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

// Destroy destroys a kind cluster
func Destroy(clName string, provider *kind.Provider) error {
	log.Infof("Deleting cluster %q ...\n", clName)
	if err := provider.Delete(clName, ""); err != nil {
		return errors.Wrapf(err, "failed to delete cluster %s", clName)
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}

	_ = os.Remove(filepath.Join(config.KindConfigDir, "kind-config-"+clName+".yaml"))
	_ = os.Remove(filepath.Join(config.LocalKubeConfigDir, "kind-config-"+clName))
	_ = os.Remove(filepath.Join(config.ContainerKubeConfigDir, "kind-config-"+clName))
	_ = os.RemoveAll(filepath.Join(usr.HomeDir, ".kube", strings.Join([]string{"kind-config", clName}, "-")))

	return nil
}
