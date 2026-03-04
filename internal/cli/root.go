package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func MustExecute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "restless",
		Short: "Terminal‑first API exploration tool",
	}

	cmd.Version = version

	// core commands
	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewGraphCmd())
	cmd.AddCommand(NewCompletionCmd(cmd))

	return cmd
}