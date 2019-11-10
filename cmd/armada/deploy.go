package armada

import (
	"io/ioutil"
	"strings"

	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/dimaunx/armada/pkg/constants"
	"github.com/dimaunx/armada/pkg/utils"
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
			box := packr.New("manifests", "../../configs")

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

			configFiles, err := ioutil.ReadDir(constants.KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range configFiles {
				clName := strings.Split(file.Name(), "-")[0]
				cl := &cluster.Cluster{Name: clName}

				kubeConfigFilePath, err := utils.GetKubeConfigPath(cl)
				if err != nil {
					log.Fatalf("%s %s", cl.Name, err)
				}

				err = utils.DeployResources(cl, netshootDeploymentFile.String(), kubeConfigFilePath, "Netshoot")
				if err != nil {
					log.Fatalf("%s %s", cl.Name, err)
				}

				err = utils.WaitForDaemonSet(cl, kubeConfigFilePath, "default", selector)
				if err != nil {
					log.Fatalf("%s %s", cl.Name, err)
				}
			}
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

			box := packr.New("manifests", "../configs")
			nginxDeploymentFile, err := box.Resolve("debug/nginx-demo-daemonset.yaml")
			if err != nil {
				log.Error(err)
			}

			configFiles, err := ioutil.ReadDir(constants.KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range configFiles {
				clName := strings.Split(file.Name(), "-")[0]
				cl := &cluster.Cluster{Name: clName}

				kubeConfigFilePath, err := utils.GetKubeConfigPath(cl)
				if err != nil {
					log.Fatal(err)
				}

				err = utils.DeployResources(cl, nginxDeploymentFile.String(), kubeConfigFilePath, "Nginx")
				if err != nil {
					log.Fatal(err)
				}

				err = utils.WaitForDeployment(cl, kubeConfigFilePath, "default", "nginx-demo")
				if err != nil {
					log.Fatal(err)
				}
			}
			return nil
		},
	}
	return cmd
}
