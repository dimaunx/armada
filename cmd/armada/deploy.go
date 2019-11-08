package armada

import (
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

type deployFlagpole struct {
	HostNetwork bool
}

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

func DeployNetshootCommand() *cobra.Command {
	flags := &deployFlagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "netshoot",
		Short: "Deploy netshoot pods for debugging",
		Long:  "Deploy netshoot pods for debugging",
		RunE: func(cmd *cobra.Command, args []string) error {
			var netshootDeploymentFilePath string
			box := packr.New("manifests", "../manifests")

			if flags.HostNetwork {
				netshootDeploymentFilePath = "debug/netshoot-daemonset-host.yaml"
			} else {
				netshootDeploymentFilePath = "debug/netshoot-daemonset.yaml"
			}

			netshootDeploymentFile, err := box.Resolve(netshootDeploymentFilePath)
			if err != nil {
				log.Error(err)
			}

			configFiles, err := ioutil.ReadDir(KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range configFiles {
				clName := strings.Split(file.Name(), "-")[0]
				cl := &Cluster{Name: clName}

				kubeConfigFilePath, err := cl.GetKubeConfigPath()
				if err != nil {
					log.Fatalf("%s %s", cl.Name, err)
				}
				err = cl.DeployResources(netshootDeploymentFile.String(), kubeConfigFilePath, "Netshoot")
				if err != nil {
					log.Errorf("%s %s", cl.Name, err)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&flags.HostNetwork, "host-network", false, "deploy the pods in host network mode.")
	return cmd
}

func DeployNginxDemoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "nginx-demo",
		Short: "Deploy nginx service and pods",
		Long:  "Deploy nginx service and pods",
		RunE: func(cmd *cobra.Command, args []string) error {

			box := packr.New("manifests", "../manifests")
			nginxDeploymentFile, err := box.Resolve("debug/nginx-demo-daemonset.yaml")
			if err != nil {
				log.Error(err)
			}

			configFiles, err := ioutil.ReadDir(KindConfigDir)
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range configFiles {
				clName := strings.Split(file.Name(), "-")[0]
				cl := &Cluster{Name: clName}

				kubeConfigFilePath, err := cl.GetKubeConfigPath()
				if err != nil {
					log.Fatalf("%s %s", cl.Name, err)
				}

				err = cl.DeployResources(nginxDeploymentFile.String(), kubeConfigFilePath, "Nginx")
				if err != nil {
					log.Errorf("%s %s", cl.Name, err)
				}
			}
			return nil
		},
	}
	return cmd
}
