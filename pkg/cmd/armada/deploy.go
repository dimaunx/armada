package armada

import (
	"io/ioutil"
	"strings"
	"sync"

	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/dimaunx/armada/pkg/deploy"
	"github.com/dimaunx/armada/pkg/wait"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DeployFlagpole struct {
	HostNetwork bool
	Clusters    []string
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
	flags := &DeployFlagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "netshoot",
		Short: "Deploy netshoot pods for debugging",
		Long:  "Deploy netshoot pods for debugging",
		RunE: func(cmd *cobra.Command, args []string) error {
			var netshootDeploymentFilePath string
			var selector string

			box := packr.New("configs", "../../../configs")

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

			var targetClusters []string
			if len(flags.Clusters) > 0 {
				targetClusters = append(targetClusters, flags.Clusters...)
			} else {

				configFiles, err := ioutil.ReadDir(config.KindConfigDir)
				if err != nil {
					log.Fatal(err)
				}

				for _, configFile := range configFiles {
					clName := strings.FieldsFunc(configFile.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
					targetClusters = append(targetClusters, clName)
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(targetClusters))
			for _, clName := range targetClusters {
				go func(clName string) {
					kubeConfigFilePath, err := cluster.GetKubeConfigPath(clName)
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}

					kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}

					clientSet, err := kubernetes.NewForConfig(kconfig)
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}

					err = deploy.Resources(clName, clientSet, netshootDeploymentFile.String(), "Netshoot")
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}

					err = wait.ForDaemonSetReady(clName, clientSet, "default", selector)
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}
					wg.Done()
				}(clName)
			}
			wg.Wait()
			return nil
		},
	}
	cmd.Flags().BoolVar(&flags.HostNetwork, "host-network", false, "deploy the pods in host network mode.")
	cmd.Flags().StringSliceVarP(&flags.Clusters, "cluster", "c", []string{}, "comma separated list of cluster names to deploy to. eg: cl1,cl6,cl3")
	return cmd
}

// DeployNginxDemoCommand returns a new cobra.Command under deploy command for armada
func DeployNginxDemoCommand() *cobra.Command {
	flags := &DeployFlagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "nginx-demo",
		Short: "Deploy nginx demo application service and pods",
		Long:  "Deploy nginx demo application service and pods",
		RunE: func(cmd *cobra.Command, args []string) error {

			box := packr.New("configs", "../../configs")

			nginxDeploymentFile, err := box.Resolve("debug/nginx-demo-daemonset.yaml")
			if err != nil {
				log.Error(err)
			}

			var targetClusters []string
			if len(flags.Clusters) > 0 {
				targetClusters = append(targetClusters, flags.Clusters...)
			} else {

				configFiles, err := ioutil.ReadDir(config.KindConfigDir)
				if err != nil {
					log.Fatal(err)
				}

				for _, configFile := range configFiles {
					clName := strings.FieldsFunc(configFile.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
					targetClusters = append(targetClusters, clName)
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(targetClusters))
			for _, clName := range targetClusters {
				go func(clName string) {
					kubeConfigFilePath, err := cluster.GetKubeConfigPath(clName)
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}

					kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}

					clientSet, err := kubernetes.NewForConfig(kconfig)
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}

					err = deploy.Resources(clName, clientSet, nginxDeploymentFile.String(), "Nginx")
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}

					err = wait.ForDaemonSetReady(clName, clientSet, "default", "nginx-demo")
					if err != nil {
						log.Fatalf("%s %s", clName, err)
					}
					wg.Done()
				}(clName)
			}
			wg.Wait()
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&flags.Clusters, "cluster", "c", []string{}, "comma separated list of cluster names to deploy to. eg: cl1,cl6,cl3")
	return cmd
}
