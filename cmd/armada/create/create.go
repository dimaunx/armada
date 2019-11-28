package create

import (
	"github.com/dimaunx/armada/cmd/armada/create/cluster"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

// CreateCmd returns a new cobra.Command under the root command for armada
func CreateCmd(provider *kind.Provider, box *packr.Box) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "create",
		Short: "Creates e2e environment",
		Long:  "Creates multiple kind based clusters",
	}

	cmd.AddCommand(cluster.CreateClustersCommand(provider, box))
	return cmd
}
