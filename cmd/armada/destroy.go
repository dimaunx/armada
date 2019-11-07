package armada

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/cluster"
)

func DestroyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "destroy",
		Short: "Destroys e2e environment",
		Long:  "Destroys multiple kind clusters",
	}
	cmd.AddCommand(DestroyClustersCommand())
	return cmd
}

func DestroyClustersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "clusters",
		Short: "Destroy e2e environment",
		Long:  "Destroys e2e environment",
		RunE: func(cmd *cobra.Command, args []string) error {

			customFormatter := new(log.TextFormatter)
			customFormatter.TimestampFormat = "2006-01-02 15:04:05"
			log.SetFormatter(customFormatter)
			customFormatter.FullTimestamp = true

			configFiles, err := ioutil.ReadDir(KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range configFiles {
				clName := strings.Split(file.Name(), "-")[0]
				cl := &Cluster{Name: clName}
				err := deleteCluster(cl)
				if err != nil {
					log.Error(err)
				}

				_ = os.Remove(filepath.Join(KindConfigDir, file.Name()))
				_ = os.Remove(filepath.Join(LocalKubeConfigDir, "kind-config-"+clName))
				_ = os.Remove(filepath.Join(ContainerKubeConfigDir, "kind-config-"+clName))
			}
			return nil
		},
	}
	return cmd
}

func deleteCluster(cl *Cluster) error {
	log.Infof("Deleting cluster %q ...\n", cl.Name)
	ctx := cluster.NewContext(cl.Name)
	if err := ctx.Delete(); err != nil {
		return errors.Wrapf(err, "failed to delete cluster %s", cl.Name)
	}
	return nil
}
