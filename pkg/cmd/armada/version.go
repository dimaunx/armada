package armada

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Build and Version
var (
	Build   string
	Version string
)

// VersionCmd returns a new cobra.Command that displays version and build information
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
