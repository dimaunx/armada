package armada

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Flags struct {
	LogLevel string
}

var (
	Build   string
	Version string
)

// NewRootCmd returns a new cobra.Command implementing the root command for armada
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "armada",
		Short: "Armada is a tool for e2e environment creation for submariner-io org",
		Long:  "Creates multiple kind clusters and e2e environments",
	}
	// add all top level subcommands
	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(DestroyCmd())
	cmd.AddCommand(DeployCmd())
	cmd.AddCommand(VersionCmd())
	return cmd
}

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "version",
		Short: "Display version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("Version: %s, Build from commit: %v", Version, Build)
			return nil
		},
	}
	return cmd
}

// Run runs the `kind` root command
func Run() error {
	return NewRootCmd().Execute()
}

// Main wraps Run and sets the log formatter
func Main() {
	// let's explicitly set stdout
	if err := Run(); err != nil {
		os.Exit(1)
	}
}
