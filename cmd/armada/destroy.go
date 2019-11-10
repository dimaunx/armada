package armada

import (
	"io/ioutil"
	"strings"

	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/dimaunx/armada/pkg/constants"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// DestroyCmd returns a new cobra.Command under root command for armada
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

// DestroyClustersCommand returns a new cobra.Command under destroy command for armada
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

			configFiles, err := ioutil.ReadDir(constants.KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range configFiles {
				clName := strings.Split(file.Name(), "-")[0]
				cl := &cluster.Cluster{Name: clName}
				err := cluster.DeleteKindCluster(cl, file)
				if err != nil {
					log.Error(err)
				}
			}
			return nil
		},
	}
	return cmd
}
