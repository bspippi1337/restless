package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "v6.0.0"
	commit  = "dev"
	date    = "unknown"
)

func NewRootCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "restless",
		Short: "REST API discovery and exploration CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if maybeAutopilot(args) {
				return fmt.Errorf("")
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringP("api", "a", "", "API context")
	cmd.PersistentFlags().StringP("cache", "c", "", "cache directory")

	cmd.AddCommand(NewScanCmd())
	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewLearnCmd())
	cmd.AddCommand(NewTeachCmd())
	cmd.AddCommand(NewCallCmd())
	cmd.AddCommand(NewShellCmd())
	cmd.AddCommand(NewMapCmd())
	cmd.AddCommand(NewGraphCmd())
	cmd.AddCommand(NewInspectCmd())
	cmd.AddCommand(NewFuzzCmd())
	cmd.AddCommand(NewCouncilCmd())
	cmd.AddCommand(NewEngineCmd())
	cmd.AddCommand(NewCopilotCmd())

	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}

	cmd.SetVersionTemplate("restless {{.Version}}\n")
	cmd.Version = fmt.Sprintf("%s (%s %s)", version, commit, date)

	AddDynamicCommands(cmd)
	return cmd
}

func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
