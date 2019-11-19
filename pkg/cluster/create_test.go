package cluster

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/dimaunx/armada/pkg/config"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/gobuffalo/packr/v2"
	kind "sigs.k8s.io/kind/pkg/cluster"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster test suite")
}

var _ = Describe("Create cluster", func() {
	AfterSuite(func() {
		configFiles, _ := ioutil.ReadDir(config.KindConfigDir)

		for _, file := range configFiles {
			clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
			err := Destroy(clName)
			Ω(err).ShouldNot(HaveOccurred())
		}
		_ = os.RemoveAll("./output")
	})
	Context("unit: Default flags", func() {
		It("Should generate config with correct default values", func() {
			flags := config.Flagpole{}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(got).Should(Equal(&config.Cluster{
				Cni:                 "kindnet",
				Name:                config.ClusterNameBase + strconv.Itoa(1),
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
				KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta2",
				NumWorkers:          config.NumWorkers,
			}))
		})
	})
	Context("unit: Custom flags", func() {
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta1", func() {
			flags := config.Flagpole{
				ImageName: "kindest/node:v1.11.1",
			}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(got).Should(Equal(&config.Cluster{
				Cni:                 "kindnet",
				Name:                config.ClusterNameBase + strconv.Itoa(1),
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
				KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta1",
				NumWorkers:          config.NumWorkers,
			}))
		})
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta2", func() {
			flags := config.Flagpole{
				ImageName: "kindest/node:v1.16.3",
			}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(got).Should(Equal(&config.Cluster{
				Cni:                 "kindnet",
				Name:                config.ClusterNameBase + strconv.Itoa(1),
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
				KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta2",
				NumWorkers:          config.NumWorkers,
			}))
		})
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta2 if image name is empty", func() {
			flags := config.Flagpole{
				ImageName: "",
			}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(got).Should(Equal(&config.Cluster{
				Cni:                 "kindnet",
				Name:                config.ClusterNameBase + strconv.Itoa(1),
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
				KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta2",
				NumWorkers:          config.NumWorkers,
			}))
		})
		It("Should return error with invalid node image name", func() {
			flags := config.Flagpole{
				ImageName: "kindest/node:1.16.3",
			}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).Should(HaveOccurred())
			Expect(got).To(BeNil())
			Expect(err).NotTo(BeNil())
		})
		It("Should set Cni to weave", func() {
			flags := config.Flagpole{
				Weave: true,
			}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(got).Should(Equal(&config.Cluster{
				Cni:                 "weave",
				Name:                config.ClusterNameBase + strconv.Itoa(1),
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
				KubeAdminAPIVersion: config.KubeAdminAPIVersion,
				NumWorkers:          config.NumWorkers,
			}))
		})
		It("Should set Cni to calico", func() {
			flags := config.Flagpole{
				Calico: true,
			}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(got).Should(Equal(&config.Cluster{
				Cni:                 "calico",
				Name:                config.ClusterNameBase + strconv.Itoa(1),
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
				KubeAdminAPIVersion: config.KubeAdminAPIVersion,
				NumWorkers:          config.NumWorkers,
			}))
		})
		It("Should set Cni to flannel", func() {
			flags := config.Flagpole{
				Flannel: true,
			}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(got).Should(Equal(&config.Cluster{
				Cni:                 "flannel",
				Name:                config.ClusterNameBase + strconv.Itoa(1),
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
				KubeAdminAPIVersion: config.KubeAdminAPIVersion,
				NumWorkers:          config.NumWorkers,
			}))
		})
		It("Should create configs for 2 clusters with flannel and overlapping cidrs", func() {
			flags := config.Flagpole{
				Flannel:     true,
				Overlap:     true,
				NumClusters: 2,
			}

			var clusters []*config.Cluster
			for i := 1; i <= flags.NumClusters; i++ {
				cl, err := PopulateClusterConfig(i, &flags)
				Ω(err).ShouldNot(HaveOccurred())
				clusters = append(clusters, cl)
			}
			Expect(len(clusters)).Should(Equal(2))
			Expect(clusters).Should(Equal([]*config.Cluster{
				{
					Cni:                 "flannel",
					Name:                config.ClusterNameBase + strconv.Itoa(1),
					PodSubnet:           "10.0.0.0/14",
					ServiceSubnet:       "100.0.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
					KubeAdminAPIVersion: config.KubeAdminAPIVersion,
					NumWorkers:          config.NumWorkers,
				},
				{
					Cni:                 "flannel",
					Name:                config.ClusterNameBase + strconv.Itoa(2),
					PodSubnet:           "10.0.0.0/14",
					ServiceSubnet:       "100.0.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(2) + ".local",
					KubeAdminAPIVersion: config.KubeAdminAPIVersion,
					NumWorkers:          config.NumWorkers,
				},
			}))
		})
	})
	Context("component: Default flags", func() {
		AfterEach(func() {
			configFiles, err := ioutil.ReadDir(config.KindConfigDir)
			Ω(err).ShouldNot(HaveOccurred())

			for _, file := range configFiles {
				clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
				err := Destroy(clName)
				Ω(err).ShouldNot(HaveOccurred())
			}
		})
		It("Should create a cluster with default settings (kindnet)", func() {
			flags := config.Flagpole{
				NumClusters: 1,
			}

			box := packr.New("configs", "../../configs")

			var clusters []*config.Cluster
			for i := 1; i <= flags.NumClusters; i++ {
				cl, err := PopulateClusterConfig(i, &flags)
				Ω(err).ShouldNot(HaveOccurred())
				clusters = append(clusters, cl)
			}

			var wg sync.WaitGroup
			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					cl.NumWorkers = 0
					err := Create(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			var knownClusters []bool
			for _, cl := range clusters {
				known, err := kind.IsKnown(cl.Name)
				Ω(err).ShouldNot(HaveOccurred())
				knownClusters = append(knownClusters, known)
			}

			Expect(len(knownClusters)).Should(Equal(1))
			Expect(clusters).Should(Equal([]*config.Cluster{
				{
					Cni:                 "kindnet",
					Name:                config.ClusterNameBase + strconv.Itoa(1),
					PodSubnet:           "10.4.0.0/14",
					ServiceSubnet:       "100.1.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
					KubeAdminAPIVersion: config.KubeAdminAPIVersion,
					NumWorkers:          0,
				},
			}))
		})
	})
	Context("component: Custom flags", func() {
		AfterEach(func() {
			configFiles, err := ioutil.ReadDir(config.KindConfigDir)
			Ω(err).ShouldNot(HaveOccurred())

			for _, file := range configFiles {
				clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
				err := Destroy(clName)
				Ω(err).ShouldNot(HaveOccurred())
			}
		})
		It("Should create a cluster with custom k8s version, default cni disabled and finalize the setup with flannel", func() {
			flags := config.Flagpole{
				NumClusters: 1,
				Flannel:     true,
				ImageName:   "kindest/node:v1.14.6",
			}

			box := packr.New("configs", "../../configs")

			var clusters []*config.Cluster
			for i := 1; i <= flags.NumClusters; i++ {
				cl, err := PopulateClusterConfig(i, &flags)
				Ω(err).ShouldNot(HaveOccurred())
				clusters = append(clusters, cl)
			}

			var wg sync.WaitGroup
			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					cl.NumWorkers = 0
					err := Create(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := FinalizeSetup(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			var knownClusters []bool
			for _, cl := range clusters {
				known, err := kind.IsKnown(cl.Name)
				Ω(err).ShouldNot(HaveOccurred())
				knownClusters = append(knownClusters, known)
			}

			ctx := context.Background()
			dockerCli, err := dockerclient.NewEnvClient()
			Ω(err).ShouldNot(HaveOccurred())

			containerFilter := filters.NewArgs()
			containerFilter.Add("name", "cl1-control-plane")
			container, err := dockerCli.ContainerList(ctx, dockertypes.ContainerListOptions{
				Filters: containerFilter,
				Limit:   1,
			})
			Ω(err).ShouldNot(HaveOccurred())
			image := container[0].Image

			Expect(image).Should(Equal(flags.ImageName))
			Expect(len(knownClusters)).Should(Equal(1))
			Expect(clusters).Should(Equal([]*config.Cluster{
				{
					Cni:                 "flannel",
					Name:                config.ClusterNameBase + strconv.Itoa(1),
					PodSubnet:           "10.4.0.0/14",
					ServiceSubnet:       "100.1.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
					KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta1",
					NumWorkers:          0,
				},
			}))
		})
		It("Should create a cluster with default cni disabled and finalize the setup with calico", func() {
			flags := config.Flagpole{
				NumClusters: 1,
				Calico:      true,
			}

			box := packr.New("configs", "../../configs")

			var clusters []*config.Cluster
			for i := 1; i <= flags.NumClusters; i++ {
				cl, err := PopulateClusterConfig(i, &flags)
				Ω(err).ShouldNot(HaveOccurred())
				clusters = append(clusters, cl)
			}

			var wg sync.WaitGroup
			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					cl.NumWorkers = 0
					err := Create(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := FinalizeSetup(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			var knownClusters []bool
			for _, cl := range clusters {
				known, err := kind.IsKnown(cl.Name)
				Ω(err).ShouldNot(HaveOccurred())
				knownClusters = append(knownClusters, known)
			}

			Expect(len(knownClusters)).Should(Equal(1))
			Expect(clusters).Should(Equal([]*config.Cluster{
				{
					Cni:                 "calico",
					Name:                config.ClusterNameBase + strconv.Itoa(1),
					PodSubnet:           "10.4.0.0/14",
					ServiceSubnet:       "100.1.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
					KubeAdminAPIVersion: config.KubeAdminAPIVersion,
					NumWorkers:          0,
				},
			}))
		})
		It("Should create a cluster with default cni disabled and finalize the setup with weave", func() {
			flags := config.Flagpole{
				NumClusters: 1,
				Weave:       true,
			}

			box := packr.New("configs", "../../configs")

			var clusters []*config.Cluster
			for i := 1; i <= flags.NumClusters; i++ {
				cl, err := PopulateClusterConfig(i, &flags)
				Ω(err).ShouldNot(HaveOccurred())
				clusters = append(clusters, cl)
			}

			var wg sync.WaitGroup
			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					cl.NumWorkers = 0
					err := Create(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			wg.Add(len(clusters))
			for _, cl := range clusters {
				go func(cl *config.Cluster) {
					err := FinalizeSetup(cl, &flags, box, &wg)
					Ω(err).ShouldNot(HaveOccurred())
				}(cl)
			}
			wg.Wait()

			var knownClusters []bool
			for _, cl := range clusters {
				known, err := kind.IsKnown(cl.Name)
				Ω(err).ShouldNot(HaveOccurred())
				knownClusters = append(knownClusters, known)
			}

			Expect(len(knownClusters)).Should(Equal(1))
			Expect(clusters).Should(Equal([]*config.Cluster{
				{
					Cni:                 "weave",
					Name:                config.ClusterNameBase + strconv.Itoa(1),
					PodSubnet:           "10.4.0.0/14",
					ServiceSubnet:       "100.1.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(1) + ".local",
					KubeAdminAPIVersion: config.KubeAdminAPIVersion,
					NumWorkers:          0,
				},
			}))
		})
	})
})
