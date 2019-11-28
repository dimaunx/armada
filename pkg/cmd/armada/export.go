package armada

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dimaunx/armada/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

type ExportFlagpole struct {
	Clusters []string
}

// ExportCmd returns a new cobra.Command under root command for armada
func ExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "export",
		Short: "Export kind cluster logs",
	}

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	cmd.AddCommand(ExportLogsCommand())
	return cmd
}

// ExportLogsCommand returns a new cobra.Command under export command for armada
func ExportLogsCommand() *cobra.Command {
	flags := &DeployFlagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "logs",
		Short: "Export kind cluster logs",
		Long:  "Export kind cluster logs",
		RunE: func(cmd *cobra.Command, args []string) error {

			provider := kind.NewProvider(
				kind.ProviderWithLogger(kindcmd.NewLogger()),
			)

			// remove existing before exporting
			_ = os.RemoveAll(filepath.Join(config.KindLogsDir, config.KindLogsDir))

			var targetClusters []string
			if len(flags.Clusters) > 0 {
				targetClusters = append(targetClusters, flags.Clusters...)
			} else {
				configFiles, err := ioutil.ReadDir(config.KindConfigDir)
				if err != nil {
					log.Fatal(err)
				}
				for _, configFile := range configFiles {
					clName := strings.FieldsFunc(configFile.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
					targetClusters = append(targetClusters, clName)
				}
			}
			for _, clName := range targetClusters {
				err := provider.CollectLogs(clName, filepath.Join(config.KindLogsDir, clName))
				if err != nil {
					log.Fatalf("%s: %v", clName, err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&flags.Clusters, "cluster", "c", []string{}, "comma separated list of cluster names. eg: cluster1,cluster6,cluster3")
	return cmd
}
