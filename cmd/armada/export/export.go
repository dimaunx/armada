package export

import (
	"github.com/dimaunx/armada/cmd/armada/export/logs"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

// ExportCmd returns a new cobra.Command under root command for armada
func ExportCmd(provider *kind.Provider) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "export",
		Short: "Export kind cluster logs",
	}
	cmd.AddCommand(logs.ExportLogsCommand(provider))
	return cmd
}
