package cluster

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/gobuffalo/packr/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Context("unit: Kind config generation", func() {
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

			flags := config.Flagpole{
				ImageName: "test/test:v1.13.2",
			}

			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "v1beta1.golden")
			configPath, err := GenerateKindConfig(cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(golden))

			_ = os.RemoveAll(configPath)
		})
		It("Should generate correct kind config for cluster with k8s version higher then 1.15", func() {

			flags := config.Flagpole{
				ImageName: "test/test:v1.16.2",
			}

			currentDir, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())

			cl, err := PopulateClusterConfig(1, &flags)
			Ω(err).ShouldNot(HaveOccurred())

			box := packr.New("configs", "../../configs")

			configDir := filepath.Join(currentDir, "testdata/kind")
			gf := filepath.Join(configDir, "v1beta2.golden")
			configPath, err := GenerateKindConfig(cl, configDir, box)
			Ω(err).ShouldNot(HaveOccurred())

			golden, err := ioutil.ReadFile(gf)
			Ω(err).ShouldNot(HaveOccurred())
			actual, err := ioutil.ReadFile(configPath)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(golden))

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
			err = PrepareKubeConfig(cl.Name, gfs, "172.17.0.3")
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
			actual, err := GenerateWeaveDeploymentFile(&cl, box)
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
			actual, err := GenerateFlannelDeploymentFile(&cl, box)
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
			actual, err := GenerateCalicoDeploymentFile(&cl, box)
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

			actual := containers[0].NetworkSettings.Networks["bridge"].IPAddress

			masterIP, err := GetMasterDockerIP("cl2")
			Ω(err).ShouldNot(HaveOccurred())

			Expect(actual).Should(Equal(masterIP))
		})
	})
})
