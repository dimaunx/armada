package cluster

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

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
		It("Should generate config with correct default values", func() {
			flags := config.Flagpole{}
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
			flags := config.Flagpole{
				ImageName: "kindest/node:v1.11.1",
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
			flags := config.Flagpole{
				ImageName: "kindest/node:v1.16.3",
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
			flags := config.Flagpole{
				ImageName: "",
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
			flags := config.Flagpole{
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
			flags := config.Flagpole{
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
			flags := config.Flagpole{
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
	})
})
