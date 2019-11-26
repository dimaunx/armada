package utils_test

import (
	"context"
	"fmt"
	"github.com/dimaunx/armada/pkg/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gobuffalo/packr/v2"

	"github.com/dimaunx/armada/pkg/cmd/armada"
	"github.com/dimaunx/armada/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils test suite")
}

var _ = Describe("Utils", func() {

	AfterSuite(func() {
		_ = os.RemoveAll("./output")
	})

	Context("unit: Default flags", func() {
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
	Context("unit: Custom flags", func() {
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
				Retain:              false,
				Tiller:              false,
				NodeImageName:       "kindest/node:v1.11.1",
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
				Retain:              false,
				Tiller:              false,
				NodeImageName:       "kindest/node:v1.16.3",
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
			configPath, err := utils.GenerateKindConfig(&cl, configDir, box)
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
			configPath, err := utils.GenerateKindConfig(&cl, configDir, box)
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
			configPath, err := utils.GenerateKindConfig(&cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(golden))

			_ = os.RemoveAll(configPath)
		})
		It("Should generate correct kind config for cluster with k8s version lower then 1.15", func() {

			flags := &armada.CreateFlagpole{
				ImageName: "test/test:v1.13.2",
				Kindnet:   true,
			}

			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl, err := armada.PopulateClusterConfig(1, flags)
			Ω(err).ShouldNot(HaveOccurred())

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "v1beta1.golden")
			cl.Name = "cl5"
			cl.DNSDomain = "cl5.local"
			configPath, err := utils.GenerateKindConfig(cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(string(actual)).Should(Equal(string(golden)))

			_ = os.RemoveAll(configPath)
		})
		It("Should generate correct kind config for cluster with k8s version higher then 1.15", func() {

			flags := &armada.CreateFlagpole{
				ImageName: "test/test:v1.16.2",
				Kindnet:   true,
			}

			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl, err := armada.PopulateClusterConfig(1, flags)
			Ω(err).ShouldNot(HaveOccurred())

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "v1beta2.golden")
			cl.Name = "cl8"
			cl.DNSDomain = "cl8.local"
			configPath, err := utils.GenerateKindConfig(cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(string(actual)).Should(Equal(string(golden)))

			_ = os.RemoveAll(configPath)
		})
	})
	Context("unit: Kubeconfigs", func() {
		It("Should generate correct kube configs for local and container based deployments", func() {
			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl := config.Cluster{
				Name: "cl1",
			}

			configDir := filepath.Join(currentDir, "testdata/kube")
			kindKubeFileName := strings.Join([]string{"kind-config", cl.Name}, "-")
			newLocalKubeFilePath := filepath.Join(currentDir, config.LocalKubeConfigDir, kindKubeFileName)
			newContainerKubeFilePath := filepath.Join(currentDir, config.ContainerKubeConfigDir, kindKubeFileName)
			gfs := filepath.Join(configDir, "kubeconfig_source")
			err = utils.PrepareKubeConfigs(cl.Name, gfs, "172.17.0.3")
			Ω(err).ShouldNot(HaveOccurred())

			local, err := ioutil.ReadFile(newLocalKubeFilePath)
			Ω(err).ShouldNot(HaveOccurred())
			container, err := ioutil.ReadFile(newContainerKubeFilePath)
			Ω(err).ShouldNot(HaveOccurred())
			localGolden, err := ioutil.ReadFile(filepath.Join(configDir, "kubeconfig_local.golden"))
			Ω(err).ShouldNot(HaveOccurred())
			containerGolden, err := ioutil.ReadFile(filepath.Join(configDir, "kubeconfig_container.golden"))
			Ω(err).ShouldNot(HaveOccurred())

			Expect(local).Should(Equal(localGolden))
			Expect(container).Should(Equal(containerGolden))
		})
	})
	Context("unit: Cni deployment files", func() {
		It("Should generate correct weave deployment file", func() {
			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl := config.Cluster{
				Name:      "cl1",
				PodSubnet: "1.2.3.4/14",
			}

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/cni")
			actual, err := utils.GenerateWeaveDeploymentFile(&cl, box)
			Ω(err).ShouldNot(HaveOccurred())
			golden, err := ioutil.ReadFile(filepath.Join(configDir, "weave_deployment.golden"))
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(string(golden)))
		})
		It("Should generate correct flannel deployment file", func() {
			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl := config.Cluster{
				Name:      "cl1",
				PodSubnet: "1.2.3.4/8",
			}

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/cni")
			actual, err := utils.GenerateFlannelDeploymentFile(&cl, box)
			Ω(err).ShouldNot(HaveOccurred())
			golden, err := ioutil.ReadFile(filepath.Join(configDir, "flannel_deployment.golden"))
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(string(golden)))
		})
		It("Should generate correct calico deployment file", func() {
			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl := config.Cluster{
				PodSubnet: "1.2.3.4/16",
			}

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/cni")
			actual, err := utils.GenerateCalicoDeploymentFile(&cl, box)
			Ω(err).ShouldNot(HaveOccurred())
			golden, err := ioutil.ReadFile(filepath.Join(configDir, "calico_deployment.golden"))
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(string(golden)))
		})
	})
	Context("component: Containers", func() {
		AfterEach(func() {
			ctx := context.Background()
			dockerCli, err := dockerclient.NewEnvClient()
			Ω(err).ShouldNot(HaveOccurred())

			containerFilter := filters.NewArgs()
			containerFilter.Add("name", "cl2-control-plane")

			containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{
				Filters: containerFilter,
				Limit:   1,
			})
			Ω(err).ShouldNot(HaveOccurred())

			err = dockerCli.ContainerRemove(ctx, containers[0].ID, types.ContainerRemoveOptions{
				Force: true,
			})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("Should return the correct ip of a master node by name", func() {

			ctx := context.Background()
			dockerCli, err := dockerclient.NewEnvClient()
			Ω(err).ShouldNot(HaveOccurred())

			reader, err := dockerCli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
			Ω(err).ShouldNot(HaveOccurred())
			_, err = io.Copy(os.Stdout, reader)
			Ω(err).ShouldNot(HaveOccurred())

			resp, err := dockerCli.ContainerCreate(ctx, &container.Config{
				Image: "alpine",
				Cmd:   []string{"/bin/sh"},
			}, nil, nil, "cl2-control-plane")
			Ω(err).ShouldNot(HaveOccurred())

			err = dockerCli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
			Ω(err).ShouldNot(HaveOccurred())

			containerFilter := filters.NewArgs()
			containerFilter.Add("name", "cl2-control-plane")

			containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{
				Filters: containerFilter,
				Limit:   1,
			})
			Ω(err).ShouldNot(HaveOccurred())

			fmt.Print(containers)
			actual := containers[0].NetworkSettings.Networks["bridge"].IPAddress

			masterIP, err := utils.GetMasterDockerIP("cl2")
			Ω(err).ShouldNot(HaveOccurred())
			fmt.Printf("actual: %s , returned: %s", actual, masterIP)

			Expect(actual).Should(Equal(masterIP))
		})
	})
})
