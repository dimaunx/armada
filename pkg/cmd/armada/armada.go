package armada

import (
	"github.com/spf13/cobra"
)

// Flags loglevel
type Flags struct {
	LogLevel string
}

// NewRootCmd returns a new cobra.Command implementing the root command for armada
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "armada",
		Short: "Armada is a tool for e2e environment creation for submariner-io org",
		Long:  "Creates multiple kind clusters and e2e environments",
	}
	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(DestroyCmd())
	cmd.AddCommand(DeployCmd())
	cmd.AddCommand(VersionCmd())
	return cmd
}
