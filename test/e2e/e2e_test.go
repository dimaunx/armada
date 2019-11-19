package e2e

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"

	"github.com/gobuffalo/packr/v2"

	"github.com/dimaunx/armada/cmd/armada"
	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/dimaunx/armada/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E test suite")
}

var _ = Describe("Cluster", func() {
	AfterSuite(func() {
		configFiles, err := ioutil.ReadDir(config.KindConfigDir)
		Ω(err).ShouldNot(HaveOccurred())

		for _, file := range configFiles {
			clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
			err := cluster.Destroy(clName)
			Ω(err).ShouldNot(HaveOccurred())
		}
		_ = os.RemoveAll("./output")
	})
	Context("e2e: Cluster creation", func() {
		var activeClusters []*config.Cluster
		It("Should create 2 clusters with calico and overlapping cidrs", func() {
			flags := config.Flagpole{
				NumClusters: 2,
				Calico:      true,
				Overlap:     true,
			}

			box := packr.New("configs", "../../configs")

			var clusters []*config.Cluster
			for i := 1; i <= flags.NumClusters; i++ {
				clName := config.ClusterNameBase + strconv.Itoa(i)
				known, err := kind.IsKnown(clName)
				Ω(err).ShouldNot(HaveOccurred())
				if known {
					log.Infof("✔ Cluster with the name %q already exists.", clName)
				} else {
					cl, err := cluster.PopulateClusterConfig(i, &flags)
					Ω(err).ShouldNot(HaveOccurred())
					clusters = append(clusters, cl)
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := cluster.Create(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := cluster.FinalizeSetup(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
					activeClusters = append(activeClusters, cl)
				}(cl)
			}
			wg.Wait()

			Expect(len(activeClusters)).Should(Equal(2))
			Expect(clusters).Should(Equal([]*config.Cluster{
				{
					Cni:                 "calico",
					Name:                config.ClusterNameBase + strconv.Itoa(1),
					PodSubnet:           "10.0.0.0/14",
					ServiceSubnet:       "100.0.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
					KubeAdminAPIVersion: config.KubeAdminAPIVersion,
					NumWorkers:          config.NumWorkers,
				},
				{
					Cni:                 "calico",
					Name:                config.ClusterNameBase + strconv.Itoa(2),
					PodSubnet:           "10.0.0.0/14",
					ServiceSubnet:       "100.0.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(2) + ".local",
					KubeAdminAPIVersion: config.KubeAdminAPIVersion,
					NumWorkers:          config.NumWorkers,
				},
			}))
		})
		It("Should create a third clusters with weave, kindest/node:v1.14.6 and tiller", func() {
			flags := config.Flagpole{
				NumClusters: 3,
				Weave:       true,
				Tiller:      true,
				ImageName:   "kindest/node:v1.14.6",
			}

			box := packr.New("configs", "../../configs")

			var clusters []*config.Cluster
			for i := 1; i <= flags.NumClusters; i++ {
				clName := config.ClusterNameBase + strconv.Itoa(i)
				known, err := kind.IsKnown(clName)
				Ω(err).ShouldNot(HaveOccurred())
				if known {
					log.Infof("✔ Cluster with the name %q already exists.", clName)
				} else {
					cl, err := cluster.PopulateClusterConfig(i, &flags)
					Ω(err).ShouldNot(HaveOccurred())
					clusters = append(clusters, cl)
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := cluster.Create(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := cluster.FinalizeSetup(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
					activeClusters = append(activeClusters, cl)
				}(cl)
			}
			wg.Wait()

			ctx := context.Background()
			dockerCli, err := dockerclient.NewEnvClient()
			Ω(err).ShouldNot(HaveOccurred())

			containerFilter := filters.NewArgs()
			containerFilter.Add("name", config.ClusterNameBase+strconv.Itoa(3)+"-control-plane")
			container, err := dockerCli.ContainerList(ctx, dockertypes.ContainerListOptions{
				Filters: containerFilter,
				Limit:   1,
			})
			Ω(err).ShouldNot(HaveOccurred())
			image := container[0].Image

			Expect(image).Should(Equal(flags.ImageName))
			Expect(len(activeClusters)).Should(Equal(3))
			Expect(clusters).Should(Equal([]*config.Cluster{
				{
					Cni:                 "weave",
					Name:                config.ClusterNameBase + strconv.Itoa(3),
					PodSubnet:           "10.12.0.0/14",
					ServiceSubnet:       "100.3.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(3) + ".local",
					KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta1",
					NumWorkers:          config.NumWorkers,
				},
			}))
		})
		It("Should not create a new cluster", func() {
			flags := config.Flagpole{
				NumClusters: 3,
			}

			for i := 1; i <= flags.NumClusters; i++ {
				clName := config.ClusterNameBase + strconv.Itoa(i)
				known, err := kind.IsKnown(clName)
				Ω(err).ShouldNot(HaveOccurred())
				if known {
					log.Infof("✔ Cluster with the name %q already exists.", clName)
				} else {
					Fail("Attempted to create a new cluster, but should have skipped as cluster already exists")
				}
			}
		})
	})
	Context("e2e: Cluster deletion", func() {
		It("Should destroy clusters 1 and 3 only", func() {
			flags := armada.DestroyFlagpole{
				Clusters: []string{config.ClusterNameBase + strconv.Itoa(1), config.ClusterNameBase + strconv.Itoa(3)},
			}

			for _, clName := range flags.Clusters {
				known, err := kind.IsKnown(clName)
				Ω(err).ShouldNot(HaveOccurred())
				if known {
					err := cluster.Destroy(clName)
					Ω(err).ShouldNot(HaveOccurred())
				}
			}

			cl1Status, err := kind.IsKnown(config.ClusterNameBase + strconv.Itoa(1))
			Ω(err).ShouldNot(HaveOccurred())
			cl2Status, err := kind.IsKnown(config.ClusterNameBase + strconv.Itoa(2))
			Ω(err).ShouldNot(HaveOccurred())
			cl3Status, err := kind.IsKnown(config.ClusterNameBase + strconv.Itoa(3))
			Ω(err).ShouldNot(HaveOccurred())

			Expect(cl1Status).Should(BeFalse())
			Expect(cl2Status).Should(BeTrue())
			Expect(cl3Status).Should(BeFalse())
		})
		It("Should destroy all remaining clusters", func() {
			configFiles, err := ioutil.ReadDir(config.KindConfigDir)
			Ω(err).ShouldNot(HaveOccurred())

			for _, file := range configFiles {
				clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
				err := cluster.Destroy(clName)
				Ω(err).ShouldNot(HaveOccurred())
			}

			cl1Status, err := kind.IsKnown(config.ClusterNameBase + strconv.Itoa(1))
			Ω(err).ShouldNot(HaveOccurred())
			cl2Status, err := kind.IsKnown(config.ClusterNameBase + strconv.Itoa(2))
			Ω(err).ShouldNot(HaveOccurred())
			cl3Status, err := kind.IsKnown(config.ClusterNameBase + strconv.Itoa(3))
			Ω(err).ShouldNot(HaveOccurred())

			Expect(cl1Status).Should(BeFalse())
			Expect(cl2Status).Should(BeFalse())
			Expect(cl3Status).Should(BeFalse())
		})
	})
})
