package armada_test

import (
	"os/user"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/dimaunx/armada/pkg/cmd/armada"
	"github.com/dimaunx/armada/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPopulateClusterConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster command test suite")
}

var _ = Describe("Populate cluster config", func() {
	Context("Default flags", func() {
		It("Should populate config with correct default values", func() {
			flags := &armada.CreateFlagpole{
				Kindnet: true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
			got, err := armada.PopulateClusterConfig(1, flags)
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
	Context("Custom flags", func() {
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta1", func() {
			flags := &armada.CreateFlagpole{
				ImageName: "kindest/node:v1.11.1",
				Kindnet:   true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
			got, err := armada.PopulateClusterConfig(1, flags)
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
				NodeImageName:       "kindest/node:v1.11.1",
				Retain:              false,
				Tiller:              false,
			}))
		})
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta2", func() {
			flags := &armada.CreateFlagpole{
				ImageName: "kindest/node:v1.16.3",
				Kindnet:   true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
			got, err := armada.PopulateClusterConfig(1, flags)
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
				NodeImageName:       "kindest/node:v1.16.3",
				Retain:              false,
				Tiller:              false,
			}))
		})
		It("Should set KubeAdminAPIVersion to kubeadm.k8s.io/v1beta2 if image name is empty", func() {
			flags := &armada.CreateFlagpole{
				ImageName: "",
				Kindnet:   true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
			got, err := armada.PopulateClusterConfig(1, flags)
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
			flags := &armada.CreateFlagpole{
				ImageName: "kindest/node:1.16.3",
			}
			got, err := armada.PopulateClusterConfig(1, flags)
			Ω(err).Should(HaveOccurred())
			Expect(got).To(BeNil())
			Expect(err).NotTo(BeNil())
		})
		It("Should set Cni to weave", func() {
			flags := &armada.CreateFlagpole{
				Weave: true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
			got, err := armada.PopulateClusterConfig(1, flags)
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
			flags := &armada.CreateFlagpole{
				Calico: true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
			got, err := armada.PopulateClusterConfig(1, flags)
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
			flags := &armada.CreateFlagpole{
				Flannel: true,
			}
			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())
			got, err := armada.PopulateClusterConfig(1, flags)
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
			flags := &armada.CreateFlagpole{
				Flannel:     true,
				Overlap:     true,
				NumClusters: 2,
			}

			usr, err := user.Current()
			Ω(err).ShouldNot(HaveOccurred())

			var clusters []*config.Cluster
			for i := 1; i <= flags.NumClusters; i++ {
				cl, err := armada.PopulateClusterConfig(i, flags)
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
