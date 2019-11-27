package armada

import (
	"io/ioutil"
	"strings"

	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/dimaunx/armada/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

// DestroyFlagpole flags for destroy command
type DestroyFlagpole struct {
	Clusters []string
}

// DestroyCmd returns a new cobra.Command under root command for armada
func DestroyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "destroy",
		Short: "Destroys e2e environment",
		Long:  "Destroys multiple kind clusters",
	}

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	provider := kind.NewProvider(
		kind.ProviderWithLogger(kindcmd.NewLogger()),
	)

	cmd.AddCommand(DestroyClustersCommand(provider))
	return cmd
}

// DestroyClustersCommand returns a new cobra.Command under destroy command for armada
func DestroyClustersCommand(provider *kind.Provider) *cobra.Command {
	flags := &DestroyFlagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "clusters",
		Short: "Destroy clusters",
		Long:  "Destroys clusters",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(flags.Clusters) > 0 {
				for _, clName := range flags.Clusters {
					known, err := cluster.IsKnown(clName, provider)
					if err != nil {
						log.Fatalf("%s: %v", clName, err)
					}
					if known {
						err := cluster.Destroy(clName, provider)
						if err != nil {
							log.Fatalf("%s: %v", clName, err)
						}
					} else {
						log.Errorf("cluster %q not found.", clName)
					}
				}
			} else {
				configFiles, err := ioutil.ReadDir(config.KindConfigDir)
				if err != nil {
					log.Fatal(err)
				}

				for _, file := range configFiles {
					clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
					err := cluster.Destroy(clName, provider)
					if err != nil {
						log.Fatalf("%s: %v", clName, err)
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&flags.Clusters, "cluster", "c", []string{}, "comma separated list of cluster names to destroy. eg: cl1,cl6,cl3")
	return cmd
}
