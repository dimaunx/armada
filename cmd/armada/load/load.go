package load

import (
	"github.com/dimaunx/armada/cmd/armada/load/image"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

// LoadCmd returns a new cobra.Command under root command for armada
func LoadCmd(provider *kind.Provider) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "load",
		Short: "Load resources in to hte cluster",
		Long:  "Load resources in to hte cluster",
	}
	cmd.AddCommand(image.LoadImageCommand(provider))
	return cmd
}
