package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// Version info can be injected via -ldflags "-X github.com/bspippi1337/restless/internal/cli.version=vX.Y.Z"
var version = "dev"

func MustExecute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restless",
		Short: "Terminal-first API exploration tool",
	}

	cmd.Version = version

	// Core workflow
	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewMapCmd())
	cmd.AddCommand(NewGraphCmd())

	// Quality-of-life
	cmd.AddCommand(NewCompletionCmd(cmd))

	return cmd
}
