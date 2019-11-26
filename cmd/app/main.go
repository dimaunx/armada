package app

import (
	"os"

	"github.com/dimaunx/armada/pkg/cmd/armada"
)

// Run runs the `armada` root command
func Run() error {
	return armada.NewRootCmd().Execute()
}

// Main wraps Run
func Main() {
	// let's explicitly set stdout
	if err := Run(); err != nil {
		os.Exit(1)
	}
}
