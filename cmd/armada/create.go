package armada

import (
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/dimaunx/armada/pkg/constants"
	"github.com/dimaunx/armada/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

// CreateCmd returns a new cobra.Command under the root command for armada
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "create",
		Short: "Creates e2e environment",
		Long:  "Creates multiple kind clusters",
	}

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	cmd.AddCommand(CreateClustersCommand())
	return cmd
}

// CreateClustersCommand returns a new cobra.Command under create command for armada
func CreateClustersCommand() *cobra.Command {
	flags := &cluster.Flagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "clusters",
		Short: "Creates multiple kubernetes clusters",
		Long:  "Creates multiple kubernetes clusters using Docker container 'nodes'",
		RunE: func(cmd *cobra.Command, args []string) error {

			if flags.Debug {
				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)
			}

			var wg sync.WaitGroup
			wg.Add(flags.NumClusters)
			for i := 1; i <= flags.NumClusters; i++ {
				go func(i int) {
					clName := constants.ClusterNameBase + strconv.Itoa(i)
					known, err := kind.IsKnown(clName)
					if err != nil {
						log.Fatal(err)
					}
					if known {
						log.Infof("✔ Cluster with the name %q already exists", clName)
						wg.Done()
					} else {
						err := utils.CreateEnvironment(i, flags, &wg)
						if err != nil {
							defer wg.Done()
							log.Fatal(err)
						}
					}
				}(i)
			}
			wg.Wait()
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			files, err := ioutil.ReadDir(constants.KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range files {
				clName := strings.Split(file.Name(), "-")[0]
				known, err := kind.IsKnown(clName)
				if err != nil {
					log.Error(err)
				}
				if known {
					log.Debugf("✔ Cluster with the name %q already exists", clName)
				} else {
					cl := &cluster.Cluster{Name: clName}
					err = utils.PrepareKubeConfig(cl)
					if err != nil {
						log.Error(err)
					}
				}
			}
			log.Infof("✔ Kubeconfigs: export KUBECONFIG=$(echo ./output/kind-config/local-dev/kind-config-cl{1..%v} | sed 's/ /:/g')", flags.NumClusters)
		},
	}
	cmd.Flags().StringVarP(&flags.ImageName, "image", "i", "", "node docker image to use for booting the cluster")
	cmd.Flags().BoolVarP(&flags.Retain, "retain", "", true, "retain nodes for debugging when cluster creation fails")
	cmd.Flags().BoolVarP(&flags.Weave, "weave", "w", false, "deploy weave")
	cmd.Flags().BoolVarP(&flags.Tiller, "tiller", "t", false, "deploy tiller")
	cmd.Flags().BoolVarP(&flags.Calico, "calico", "c", false, "deploy calico")
	cmd.Flags().BoolVarP(&flags.Flannel, "flannel", "f", false, "deploy flannel")
	cmd.Flags().BoolVarP(&flags.Debug, "debug", "v", false, "set log level to debug")
	cmd.Flags().DurationVar(&flags.Wait, "wait", 5*time.Minute, "amount of minutes to wait for control plane nodes to be ready")
	cmd.Flags().IntVarP(&flags.NumClusters, "num", "n", 2, "number of clusters to create")
	return cmd
}
