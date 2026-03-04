package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func MustExecute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restless",
		Short: "Terminal-first API workbench",
		Long:  "Restless discovers APIs and helps you explore them quickly.",
	}

	cmd.Version = fmt.Sprintf("%s (%s) %s", version, commit, date)

	cmd.AddCommand(NewScanCmd())
	cmd.AddCommand(NewMapCmd())
	cmd.AddCommand(NewInspectCmd())
	cmd.AddCommand(NewFuzzCmd())
	cmd.AddCommand(NewReplayCmd())
	cmd.AddCommand(NewGraphCmd())
	cmd.AddCommand(NewDiscoverCmd())

	return cmd
}
