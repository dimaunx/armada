package armada

import (
	"io/ioutil"
	"net"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/pkg/errors"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

// CreateFlagpole is a list of cli flags for create clusters command
type CreateFlagpole struct {
	// ImageName is the node image used for cluster creation
	ImageName string

	// Wait is a time duration to wait until cluster is ready
	Wait time.Duration

	// Retain if you keep clusters running even if error occurs
	Retain bool

	// Weave if to install weave cni
	Weave bool

	// Flannel if to install flannel cni
	Flannel bool

	// Calico if to install calico cni
	Calico bool

	// Kindnet if to install kindnet default cni
	Kindnet bool

	// Debug if to enable debug log level
	Debug bool

	// DeployTiller if to install tiller
	Tiller bool

	// Overlap if to create clusters with overlapping cidrs
	Overlap bool

	// NumClusters is the number of clusters to create
	NumClusters int
}

// PopulateClusterConfig return a desired cluster config object
func PopulateClusterConfig(i int, flags *CreateFlagpole) (*config.Cluster, error) {

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	cl := &config.Cluster{
		Name:                config.ClusterNameBase + strconv.Itoa(i),
		NodeImageName:       flags.ImageName,
		NumWorkers:          config.NumWorkers,
		DNSDomain:           config.ClusterNameBase + strconv.Itoa(i) + ".local",
		KubeAdminAPIVersion: config.KubeAdminAPIVersion,
		Retain:              flags.Retain,
		Tiller:              flags.Tiller,
		KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", strings.Join([]string{"kind-config", config.ClusterNameBase + strconv.Itoa(i)}, "-")),
	}

	podIP := net.ParseIP(config.PodCidrBase)
	podIP = podIP.To4()
	serviceIP := net.ParseIP(config.ServiceCidrBase)
	serviceIP = serviceIP.To4()

	if !flags.Overlap {
		podIP[1] += byte(4 * i)
		serviceIP[1] += byte(i)
	}

	cl.PodSubnet = podIP.String() + config.PodCidrMask
	cl.ServiceSubnet = serviceIP.String() + config.ServiceCidrMask

	if flags.Weave {
		cl.Cni = "weave"
		cl.WaitForReady = 0
	} else if flags.Calico {
		cl.Cni = "calico"
		cl.WaitForReady = 0
	} else if flags.Flannel {
		cl.Cni = "flannel"
		cl.WaitForReady = 0
	} else if flags.Kindnet {
		cl.Cni = "kindnet"
		cl.WaitForReady = flags.Wait
	}

	if flags.ImageName != "" {
		tgt := semver.MustParse("1.15")
		results := strings.Split(flags.ImageName, ":v")
		if len(results) == 2 {
			sver := semver.MustParse(results[len(results)-1])
			if sver.LessThan(tgt) {
				cl.KubeAdminAPIVersion = "kubeadm.k8s.io/v1beta1"
			}
		} else {
			return nil, errors.Errorf("%q: Could not extract version from %s, split is by ':v', example of correct image name: kindest/node:v1.15.3.", cl.Name, flags.ImageName)
		}
	}
	return cl, nil
}

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
	flags := &CreateFlagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "clusters",
		Short: "Creates multiple kubernetes clusters",
		Long:  "Creates multiple kubernetes clusters using Docker container 'nodes'",
		RunE: func(cmd *cobra.Command, args []string) error {

			provider := kind.NewProvider(
				kind.ProviderWithLogger(kindcmd.NewLogger()),
			)

			if flags.Debug {
				log.SetLevel(log.DebugLevel)
				//log.SetReportCaller(true)
			}

			var clusters []*config.Cluster
			box := packr.New("configs", "../../../configs")
			for i := 1; i <= flags.NumClusters; i++ {
				clName := config.ClusterNameBase + strconv.Itoa(i)
				known, err := cluster.IsKnown(clName, provider)
				if err != nil {
					log.Fatalf("%s: %v", clName, err)
				}
				if known {
					log.Infof("✔ Cluster with the name %q already exists.", clName)
				} else {
					cl, err := PopulateClusterConfig(i, flags)
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
					err := cluster.Create(cl, provider, box, &wg)
					if err != nil {
						defer wg.Done()
						log.Fatalf("%s: %s", cl.Name, err)
					}
				}(cl)
			}
			wg.Wait()

			log.Info("Finalizing the clusters setup ...")
			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := cluster.FinalizeSetup(cl, box, &wg)
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
			log.Infof("✔ Kubeconfigs: export KUBECONFIG=$(echo ./%s/kind-config-%s{1..%v} | sed 's/ /:/g')", config.LocalKubeConfigDir, config.ClusterNameBase, flags.NumClusters)
		},
	}
	cmd.Flags().StringVarP(&flags.ImageName, "image", "i", "", "node docker image to use for booting the cluster")
	cmd.Flags().BoolVarP(&flags.Retain, "retain", "", true, "retain nodes for debugging when cluster creation fails")
	cmd.Flags().BoolVarP(&flags.Weave, "weave", "w", false, "deploy with weave")
	cmd.Flags().BoolVarP(&flags.Tiller, "tiller", "t", false, "deploy with tiller")
	cmd.Flags().BoolVarP(&flags.Calico, "calico", "c", false, "deploy with calico")
	cmd.Flags().BoolVarP(&flags.Kindnet, "kindnet", "k", true, "deploy with kindnet default cni")
	cmd.Flags().BoolVarP(&flags.Flannel, "flannel", "f", false, "deploy with flannel")
	cmd.Flags().BoolVarP(&flags.Debug, "debug", "v", false, "set log level to debug")
	cmd.Flags().BoolVarP(&flags.Overlap, "overlap", "o", false, "create clusters with overlapping cidrs")
	cmd.Flags().DurationVar(&flags.Wait, "wait", 5*time.Minute, "amount of minutes to wait for control plane nodes to be ready")
	cmd.Flags().IntVarP(&flags.NumClusters, "num", "n", 2, "number of clusters to create")
	return cmd
}
