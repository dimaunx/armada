package armada

import (
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/dimaunx/armada/pkg/config"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
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

	provider := kind.NewProvider(
		kind.ProviderWithLogger(kindcmd.NewLogger()),
	)

	cmd.AddCommand(CreateClustersCommand(provider))
	return cmd
}

// CreateClustersCommand returns a new cobra.Command under create command for armada
func CreateClustersCommand(provider *kind.Provider) *cobra.Command {
	flags := &config.Flagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "clusters",
		Short: "Creates multiple kubernetes clusters",
		Long:  "Creates multiple kubernetes clusters using Docker container 'nodes'",
		RunE: func(cmd *cobra.Command, args []string) error {

			if flags.Debug {
				log.SetLevel(log.DebugLevel)
				//log.SetReportCaller(true)
			}

			var clusters []*config.Cluster
			box := packr.New("configs", "../../configs")
			for i := 1; i <= flags.NumClusters; i++ {
				clName := config.ClusterNameBase + strconv.Itoa(i)
				known, err := cluster.IsKnown(clName, provider)
				if err != nil {
					log.Fatalf("%s: %v", clName, err)
				}
				if known {
					log.Infof("✔ Cluster with the name %q already exists.", clName)
				} else {
					cl, err := cluster.PopulateClusterConfig(i, flags)
					if err != nil {
						log.Fatal(err)
					}
					clusters = append(clusters, cl)
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {

					err := cluster.Create(cl, flags, provider, box, &wg)
					if err != nil {
						defer wg.Done()
						log.Fatalf("%s: %s", cl.Name, err)
					}
				}(cl)
			}
			wg.Wait()

			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := cluster.FinalizeSetup(cl, flags, box, &wg)
					if err != nil {
						defer wg.Done()
						log.Fatalf("%s: %s", cl.Name, err)
					}
				}(cl)
			}
			wg.Wait()
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			files, err := ioutil.ReadDir(config.KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			provider := kind.NewProvider()

			for _, file := range files {
				clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
				known, err := cluster.IsKnown(clName, provider)
				if err != nil {
					log.Error(err)
				}
				if !known {
					cl := &config.Cluster{Name: clName}
					usr, err := user.Current()
					if err != nil {
						log.Error(err)
					}

					kindKubeFileName := strings.Join([]string{"kind-config", cl.Name}, "-")
					kindKubeFilePath := filepath.Join(usr.HomeDir, ".kube", kindKubeFileName)

					masterIP, err := cluster.GetMasterDockerIP(clName)
					if err != nil {
						log.Error(err)
					}

					err = cluster.PrepareKubeConfigs(clName, kindKubeFilePath, masterIP)
					if err != nil {
						log.Error(err)
					}
				}
			}
			log.Infof("✔ Kubeconfigs: export KUBECONFIG=$(echo ./output/kind-config/local-dev/kind-config-%s{1..%v} | sed 's/ /:/g')", config.ClusterNameBase, flags.NumClusters)
		},
	}
	cmd.Flags().StringVarP(&flags.ImageName, "image", "i", "", "node docker image to use for booting the cluster")
	cmd.Flags().BoolVarP(&flags.Retain, "retain", "", true, "retain nodes for debugging when cluster creation fails")
	cmd.Flags().BoolVarP(&flags.Weave, "weave", "w", false, "deploy with weave")
	cmd.Flags().BoolVarP(&flags.Tiller, "tiller", "t", false, "deploy with tiller")
	cmd.Flags().BoolVarP(&flags.Calico, "calico", "c", false, "deploy with calico")
	cmd.Flags().BoolVarP(&flags.Flannel, "flannel", "f", false, "deploy with flannel")
	cmd.Flags().BoolVarP(&flags.Debug, "debug", "v", false, "set log level to debug")
	cmd.Flags().BoolVarP(&flags.Overlap, "overlap", "o", false, "create clusters with overlapping cidrs")
	cmd.Flags().DurationVar(&flags.Wait, "wait", 5*time.Minute, "amount of minutes to wait for control plane nodes to be ready")
	cmd.Flags().IntVarP(&flags.NumClusters, "num", "n", 2, "number of clusters to create")
	return cmd
}
