package cluster

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gobuffalo/packr/v2"

	"github.com/dimaunx/armada/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kind "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster test suite")
}

var _ = Describe("Create cluster", func() {

	AfterSuite(func() {

		provider := kind.NewProvider(
			kind.ProviderWithLogger(kindcmd.NewLogger()),
		)

		configFiles, _ := ioutil.ReadDir(config.KindConfigDir)

		for _, file := range configFiles {
			clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
			err := Destroy(clName, provider)
			Ω(err).ShouldNot(HaveOccurred())
		}
		_ = os.RemoveAll("./output")
	})
	Context("unit: Default flags", func() {
		It("Should populate config with correct default values", func() {
			flags := config.CreateFlagpole{
				Kindnet: true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
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
				KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(1)),
			}))
		})
	})
	Context("unit: Custom flags", func() {
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta1", func() {
			flags := config.CreateFlagpole{
				ImageName: "kindest/node:v1.11.1",
				Kindnet:   true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
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
				KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(1)),
			}))
		})
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta2", func() {
			flags := config.CreateFlagpole{
				ImageName: "kindest/node:v1.16.3",
				Kindnet:   true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
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
				KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(1)),
			}))
		})
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta2 if image name is empty", func() {
			flags := config.CreateFlagpole{
				ImageName: "",
				Kindnet:   true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
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
				KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(1)),
			}))
		})
		It("Should return error with invalid node image name", func() {
			flags := config.CreateFlagpole{
				ImageName: "kindest/node:1.16.3",
			}
			got, err := PopulateClusterConfig(1, &flags)
			Ω(err).Should(HaveOccurred())
			Expect(got).To(BeNil())
			Expect(err).NotTo(BeNil())
		})
		It("Should set Cni to weave", func() {
			flags := config.CreateFlagpole{
				Weave: true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
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
				KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(1)),
			}))
		})
		It("Should set Cni to calico", func() {
			flags := config.CreateFlagpole{
				Calico: true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
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
				KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(1)),
			}))
		})
		It("Should set Cni to flannel", func() {
			flags := config.CreateFlagpole{
				Flannel: true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
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
				KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(1)),
			}))
		})
		It("Should create configs for 2 clusters with flannel and overlapping cidrs", func() {
			flags := config.CreateFlagpole{
				Flannel:     true,
				Overlap:     true,
				NumClusters: 2,
			}

			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())

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
					KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(1)),
				},
				{
					Cni:                 "flannel",
					Name:                config.ClusterNameBase + strconv.Itoa(2),
					PodSubnet:           "10.0.0.0/14",
					ServiceSubnet:       "100.0.0.0/16",
					DNSDomain:           config.ClusterNameBase + strconv.Itoa(2) + ".local",
					KubeAdminAPIVersion: config.KubeAdminAPIVersion,
					NumWorkers:          config.NumWorkers,
					KubeConfigFilePath:  filepath.Join(usr.HomeDir, ".kube", "kind-config-"+config.ClusterNameBase+strconv.Itoa(2)),
				},
			}))
		})
		It("Should generate correct kind config for default cni", func() {
			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl := config.Cluster{
				Cni:                 "kindnet",
				Name:                "default",
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           "cl1.local",
				KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta2",
				NumWorkers:          2,
			}

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "default_cni.golden")
			configPath, err := GenerateKindConfig(&cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(golden))

			_ = os.RemoveAll(configPath)
		})
		It("Should generate correct kind config for custom cni", func() {
			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl := config.Cluster{
				Cni:                 "weave",
				Name:                "custom",
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           "cl1.local",
				KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta2",
				NumWorkers:          2,
			}

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "custom_cni.golden")
			configPath, err := GenerateKindConfig(&cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(golden))

			_ = os.RemoveAll(configPath)
		})
		It("Should generate correct kind config for cluster with 5 workers and custom cni", func() {
			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl := config.Cluster{
				Cni:                 "flannel",
				Name:                "custom",
				PodSubnet:           "10.4.0.0/14",
				ServiceSubnet:       "100.1.0.0/16",
				DNSDomain:           "cl1.local",
				KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta2",
				NumWorkers:          5,
			}

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "custom_five_workers.golden")
			configPath, err := GenerateKindConfig(&cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(golden))

			_ = os.RemoveAll(configPath)
		})
		It("Should generate correct kind config for cluster with k8s version lower then 1.15", func() {

			flags := config.CreateFlagpole{
				ImageName: "test/test:v1.13.2",
				Kindnet:   true,
			}

			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "v1beta1.golden")
			cl.Name = "cl5"
			cl.DNSDomain = "cl5.local"
			configPath, err := GenerateKindConfig(cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(string(actual)).Should(Equal(string(golden)))

			_ = os.RemoveAll(configPath)
		})
		It("Should generate correct kind config for cluster with k8s version higher then 1.15", func() {

			flags := config.CreateFlagpole{
				ImageName: "test/test:v1.16.2",
				Kindnet:   true,
			}

			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "v1beta2.golden")
			cl.Name = "cl8"
			cl.DNSDomain = "cl8.local"
			configPath, err := GenerateKindConfig(cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(string(actual)).Should(Equal(string(golden)))

			_ = os.RemoveAll(configPath)
		})
	})
})
