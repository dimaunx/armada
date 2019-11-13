package armada

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/dimaunx/armada/pkg/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/dimaunx/armada/pkg/waiter"

	"github.com/dimaunx/armada/pkg/deploy"
	"github.com/dimaunx/armada/pkg/util"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type deployFlagpole struct {
	HostNetwork bool
}

// DeployCmd returns a new cobra.Command under root command for armada
func DeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "deploy",
		Short: "Deploy resources",
		Long:  "Deploy resources",
	}

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	cmd.AddCommand(DeployNetshootCommand())
	cmd.AddCommand(DeployNginxDemoCommand())
	return cmd
}

// DeployNetshootCommand returns a new cobra.Command under deploy command for armada
func DeployNetshootCommand() *cobra.Command {
	flags := &deployFlagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "netshoot",
		Short: "Deploy netshoot pods for debugging",
		Long:  "Deploy netshoot pods for debugging",
		RunE: func(cmd *cobra.Command, args []string) error {
			var netshootDeploymentFilePath string
			var selector string

			box := packr.New("configs", "../../configs")

			if flags.HostNetwork {
				netshootDeploymentFilePath = "debug/netshoot-daemonset-host.yaml"
				selector = "netshoot-host-net"
			} else {
				netshootDeploymentFilePath = "debug/netshoot-daemonset.yaml"
				selector = "netshoot"
			}

			netshootDeploymentFile, err := box.Resolve(netshootDeploymentFilePath)
			if err != nil {
				log.Error(err)
			}

			configFiles, err := ioutil.ReadDir(config.KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			var wg sync.WaitGroup
			wg.Add(len(configFiles))
			for _, file := range configFiles {
				go func(file os.FileInfo) {
					clName := strings.Split(file.Name(), "-")[0]
					cl := &config.Cluster{Name: clName}

					kubeConfigFilePath, err := util.GetKubeConfigPath(cl)
					if err != nil {
						log.Fatalf("%s %s", cl.Name, err)
					}

					kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
					if err != nil {
						log.Fatalf("%s %s", cl.Name, err)
					}

					clientSet, err := kubernetes.NewForConfig(kconfig)
					if err != nil {
						log.Fatalf("%s %s", cl.Name, err)
					}

					err = deploy.DeployResources(cl, clientSet, netshootDeploymentFile.String(), "Netshoot")
					if err != nil {
						log.Fatalf("%s %s", cl.Name, err)
					}

					err = waiter.WaitForDaemonSet(cl, clientSet, "default", selector)
					if err != nil {
						log.Fatalf("%s %s", cl.Name, err)
					}
					wg.Done()
				}(file)
			}
			wg.Wait()
			return nil
		},
	}
	cmd.Flags().BoolVar(&flags.HostNetwork, "host-network", false, "deploy the pods in host network mode.")
	return cmd
}

// DeployNginxDemoCommand returns a new cobra.Command under deploy command for armada
func DeployNginxDemoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "nginx-demo",
		Short: "Deploy nginx service and pods",
		Long:  "Deploy nginx service and pods",
		RunE: func(cmd *cobra.Command, args []string) error {

			box := packr.New("configs", "../../configs")

			nginxDeploymentFile, err := box.Resolve("debug/nginx-demo-daemonset.yaml")
			if err != nil {
				log.Error(err)
			}

			configFiles, err := ioutil.ReadDir(config.KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range configFiles {
				clName := strings.Split(file.Name(), "-")[0]
				cl := &config.Cluster{Name: clName}

				kubeConfigFilePath, err := util.GetKubeConfigPath(cl)
				if err != nil {
					log.Fatal(err)
				}

				kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
				if err != nil {
					return err
				}

				clientSet, err := kubernetes.NewForConfig(kconfig)
				if err != nil {
					return err
				}

				err = deploy.DeployResources(cl, clientSet, nginxDeploymentFile.String(), "Nginx")
				if err != nil {
					log.Fatal(err)
				}

				err = waiter.WaitForDaemonSet(cl, clientSet, "default", "nginx-demo")
				if err != nil {
					log.Fatal(err)
				}
			}
			return nil
		},
	}
	return cmd
}
