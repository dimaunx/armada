package armada

import (
	"os"

	"github.com/dimaunx/armada/cmd/armada/load"

	"github.com/dimaunx/armada/cmd/armada/create"
	"github.com/dimaunx/armada/cmd/armada/deploy"
	"github.com/dimaunx/armada/cmd/armada/destroy"
	"github.com/dimaunx/armada/cmd/armada/export"
	"github.com/dimaunx/armada/cmd/armada/version"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

// NewRootCmd returns a new cobra.Command implementing the root command for armada
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "armada",
		Short: "Armada is a tool for e2e environment creation for submariner-io org",
		Long:  "Creates multiple kind clusters and e2e environments",
	}

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	provider := kind.NewProvider(
		kind.ProviderWithLogger(kindcmd.NewLogger()),
	)

	box := packr.New("configs", "../../configs")

	cmd.AddCommand(create.CreateCmd(provider, box))
	cmd.AddCommand(destroy.DestroyCmd(provider))
	cmd.AddCommand(export.ExportCmd(provider))
	cmd.AddCommand(load.LoadCmd(provider))
	cmd.AddCommand(deploy.DeployCmd(box))
	cmd.AddCommand(version.VersionCmd())
	return cmd
}

// Run runs the `armada` root command
func Run() error {
	return NewRootCmd().Execute()
}

// Main wraps Run
func Main() {
	// let's explicitly set stdout
	if err := Run(); err != nil {
		os.Exit(1)
	}
}
