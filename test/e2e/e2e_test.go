package e2e

import (
	"context"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	createclustercmd "github.com/dimaunx/armada/cmd/armada/create/cluster"
	deploynginxcmd "github.com/dimaunx/armada/cmd/armada/deploy/nginx"
	destroycmd "github.com/dimaunx/armada/cmd/armada/destroy/cluster"
	exportlogscmd "github.com/dimaunx/armada/cmd/armada/export/logs"
	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/dimaunx/armada/pkg/defaults"
	"github.com/dimaunx/armada/pkg/deploy"
	"github.com/dimaunx/armada/pkg/wait"
	"github.com/gobuffalo/packr/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	kind "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func CreateEnvironment(flags *createclustercmd.CreateClusterFlagpole, provider *kind.Provider) ([]*cluster.Config, error) {
	box := packr.New("configs", "../../configs")
	var clusters []*cluster.Config
	for i := 1; i <= flags.NumClusters; i++ {
		clName := defaults.ClusterNameBase + strconv.Itoa(i)
		known, err := cluster.IsKnown(clName, provider)
		if err != nil {
			return nil, err
		}
		if known {
			log.Infof("✔ Cluster with the name %q already exists.", clName)
		} else {
			cni := createclustercmd.GetCniFromFlags(flags)
			cl, err := cluster.PopulateConfig(i, flags.ImageName, cni, flags.Retain, flags.Tiller, flags.Overlap, flags.Wait)
			if err != nil {
				return nil, err
			}
			clusters = append(clusters, cl)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(clusters))
	for _, cl := range clusters {
		go func(cl *cluster.Config) {
			err := cluster.Create(cl, provider, box, &wg)
			if err != nil {
				log.Fatal(err)
			}
		}(cl)
	}
	wg.Wait()

	wg.Add(len(clusters))
	for _, cl := range clusters {
		go func(cl *cluster.Config) {
			err := cluster.FinalizeSetup(cl, box, &wg)
			if err != nil {
				log.Fatal(err)
			}
		}(cl)
	}
	wg.Wait()
	return clusters, nil
}

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E test suite")
}

var _ = Describe("E2E Tests", func() {

	provider := kind.NewProvider(
		kind.ProviderWithLogger(kindcmd.NewLogger()),
	)

	var _ = AfterSuite(func() {
		_ = os.RemoveAll("./output")
	})

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	Context("Cluster creation and deployment", func() {
		It("Should create 2 clusters with flannel and overlapping cidrs", func() {
			flags := &createclustercmd.CreateClusterFlagpole{
				NumClusters: 2,
				Overlap:     true,
				Flannel:     true,
				Retain:      false,
				Wait:        5 * time.Minute,
			}

			clusters, err := CreateEnvironment(flags, provider)
			Ω(err).ShouldNot(HaveOccurred())

			cl1Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(1), provider)
			Ω(err).ShouldNot(HaveOccurred())
			cl2Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(2), provider)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(cl1Status).Should(BeTrue())
			Expect(cl2Status).Should(BeTrue())
			Expect(clusters).Should(Equal([]*cluster.Config{
				{
					Cni:                 "flannel",
					Name:                defaults.ClusterNameBase + strconv.Itoa(1),
					PodSubnet:           "10.0.0.0/14",
					ServiceSubnet:       "100.0.0.0/16",
					DNSDomain:           defaults.ClusterNameBase + strconv.Itoa(1) + ".local",
					KubeAdminAPIVersion: defaults.KubeAdminAPIVersion,
					NumWorkers:          defaults.NumWorkers,
					KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+defaults.ClusterNameBase+strconv.Itoa(1)),
					Retain:              false,
					WaitForReady:        0,
				},
				{
					Cni:                 "flannel",
					Name:                defaults.ClusterNameBase + strconv.Itoa(2),
					PodSubnet:           "10.0.0.0/14",
					ServiceSubnet:       "100.0.0.0/16",
					DNSDomain:           defaults.ClusterNameBase + strconv.Itoa(2) + ".local",
					KubeAdminAPIVersion: defaults.KubeAdminAPIVersion,
					NumWorkers:          defaults.NumWorkers,
					KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+defaults.ClusterNameBase+strconv.Itoa(2)),
					Retain:              false,
					WaitForReady:        0,
				},
			}))
		})
		It("Should create a third cluster with weave, kindest/node:v1.15.6 and tiller", func() {
			flags := &createclustercmd.CreateClusterFlagpole{
				NumClusters: 3,
				Weave:       true,
				Tiller:      true,
				ImageName:   "kindest/node:v1.15.6",
				Retain:      false,
				Wait:        5 * time.Minute,
			}

			clusters, err := CreateEnvironment(flags, provider)
			Ω(err).ShouldNot(HaveOccurred())

			ctx := context.Background()
			dockerCli, err := dockerclient.NewEnvClient()
			Ω(err).ShouldNot(HaveOccurred())

			containerFilter := filters.NewArgs()
			containerFilter.Add("name", defaults.ClusterNameBase+strconv.Itoa(3)+"-control-plane")
			container, err := dockerCli.ContainerList(ctx, dockertypes.ContainerListOptions{
				Filters: containerFilter,
				Limit:   1,
			})
			Ω(err).ShouldNot(HaveOccurred())
			image := container[0].Image
			cl3Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(3), provider)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(image).Should(Equal(flags.ImageName))
			Expect(cl3Status).Should(BeTrue())
			Expect(clusters).Should(Equal([]*cluster.Config{
				{
					Cni:                 "weave",
					Name:                defaults.ClusterNameBase + strconv.Itoa(3),
					PodSubnet:           "10.12.0.0/14",
					ServiceSubnet:       "100.3.0.0/16",
					DNSDomain:           defaults.ClusterNameBase + strconv.Itoa(3) + ".local",
					KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta2",
					NumWorkers:          defaults.NumWorkers,
					KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+defaults.ClusterNameBase+strconv.Itoa(3)),
					WaitForReady:        0,
					NodeImageName:       "kindest/node:v1.15.6",
					Retain:              false,
					Tiller:              true,
				},
			}))
		})
		It("Should deploy nginx-demo to clusters 1 and 3", func() {

			flags := &deploynginxcmd.NginxDeployFlagpole{
				Clusters: []string{"cluster1", "cluster3"},
				Debug:    true,
			}

			box := packr.New("configs", "../../configs")
			nginxDeploymentFile, err := box.Resolve("debug/nginx-demo-daemonset.yaml")
			Ω(err).ShouldNot(HaveOccurred())

			var activeDeployments []string
			var wg sync.WaitGroup
			wg.Add(len(flags.Clusters))
			for _, clName := range flags.Clusters {
				go func(clName string) {
					kubeConfigFilePath, err := cluster.GetKubeConfigPath(clName)
					Ω(err).ShouldNot(HaveOccurred())

					kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
					Ω(err).ShouldNot(HaveOccurred())

					clientSet, err := kubernetes.NewForConfig(kconfig)
					Ω(err).ShouldNot(HaveOccurred())

					err = deploy.Resources(clName, clientSet, nginxDeploymentFile.String(), "Nginx")
					Ω(err).ShouldNot(HaveOccurred())

					err = wait.ForDaemonSetReady(clName, clientSet, "default", "nginx-demo")
					Ω(err).ShouldNot(HaveOccurred())
					activeDeployments = append(activeDeployments, clName)
					wg.Done()
				}(clName)
			}
			wg.Wait()

			Expect(len(activeDeployments)).Should(Equal(2))

		})
		It("Should deploy netshoot to all 3 clusters", func() {
			box := packr.New("configs", "../../configs")
			netshootDeploymentFile, err := box.Resolve("debug/netshoot-daemonset.yaml")
			Ω(err).ShouldNot(HaveOccurred())

			configFiles, err := ioutil.ReadDir(defaults.KindConfigDir)
			Ω(err).ShouldNot(HaveOccurred())

			var activeDeployments []string
			var wg sync.WaitGroup
			wg.Add(len(configFiles))
			for _, file := range configFiles {
				go func(file os.FileInfo) {
					clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
					kubeConfigFilePath, err := cluster.GetKubeConfigPath(clName)
					Ω(err).ShouldNot(HaveOccurred())

					kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
					Ω(err).ShouldNot(HaveOccurred())

					clientSet, err := kubernetes.NewForConfig(kconfig)
					Ω(err).ShouldNot(HaveOccurred())

					err = deploy.Resources(clName, clientSet, netshootDeploymentFile.String(), "Netshoot")
					Ω(err).ShouldNot(HaveOccurred())

					err = wait.ForDaemonSetReady(clName, clientSet, "default", "netshoot")
					Ω(err).ShouldNot(HaveOccurred())
					activeDeployments = append(activeDeployments, clName)
					wg.Done()
				}(file)
			}
			wg.Wait()

			Expect(len(configFiles)).Should(Equal(3))
			Expect(len(activeDeployments)).Should(Equal(3))
		})
		It("Should not create a new cluster", func() {
			flags := &createclustercmd.CreateClusterFlagpole{
				NumClusters: 3,
			}

			for i := 1; i <= flags.NumClusters; i++ {
				clName := defaults.ClusterNameBase + strconv.Itoa(i)
				known, err := cluster.IsKnown(clName, provider)
				Ω(err).ShouldNot(HaveOccurred())
				if known {
					log.Infof("✔ Cluster with the name %q already exists.", clName)
				} else {
					Fail("Attempted to create a new cluster, but should have skipped as cluster already exists")
				}
			}
		})
		It("Should export logs for clusters 1 and 2", func() {
			flags := &exportlogscmd.ExportLogsFlagpole{
				Clusters: []string{"cluster1", "cluster2"},
			}

			for _, clName := range flags.Clusters {
				err := provider.CollectLogs(clName, filepath.Join(defaults.KindLogsDir, clName))
				Ω(err).ShouldNot(HaveOccurred())
			}

			_, err := os.Stat(filepath.Join(defaults.KindLogsDir, "cluster1", "cluster1-control-plane"))
			Ω(err).ShouldNot(HaveOccurred())
			_, err = os.Stat(filepath.Join(defaults.KindLogsDir, "cluster2", "cluster2-control-plane"))
			Ω(err).ShouldNot(HaveOccurred())

		})
	})
	Context("Config deletion", func() {
		It("Should destroy clusters 1 and 3 only", func() {
			flags := destroycmd.DestroyClusterFlagpole{
				Clusters: []string{defaults.ClusterNameBase + strconv.Itoa(1), defaults.ClusterNameBase + strconv.Itoa(3)},
			}

			for _, clName := range flags.Clusters {
				known, err := cluster.IsKnown(clName, provider)
				Ω(err).ShouldNot(HaveOccurred())
				if known {
					err := cluster.Destroy(clName, provider)
					Ω(err).ShouldNot(HaveOccurred())
				}
			}

			cl1Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(1), provider)
			Ω(err).ShouldNot(HaveOccurred())
			cl2Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(2), provider)
			Ω(err).ShouldNot(HaveOccurred())
			cl3Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(3), provider)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(cl1Status).Should(BeFalse())
			Expect(cl2Status).Should(BeTrue())
			Expect(cl3Status).Should(BeFalse())
		})
		It("Should destroy all remaining clusters", func() {
			configFiles, err := ioutil.ReadDir(defaults.KindConfigDir)
			Ω(err).ShouldNot(HaveOccurred())

			for _, file := range configFiles {
				clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
				err := cluster.Destroy(clName, provider)
				Ω(err).ShouldNot(HaveOccurred())
			}

			cl1Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(1), provider)
			Ω(err).ShouldNot(HaveOccurred())
			cl2Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(2), provider)
			Ω(err).ShouldNot(HaveOccurred())
			cl3Status, err := cluster.IsKnown(defaults.ClusterNameBase+strconv.Itoa(3), provider)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(cl1Status).Should(BeFalse())
			Expect(cl2Status).Should(BeFalse())
			Expect(cl3Status).Should(BeFalse())
		})
	})
})
