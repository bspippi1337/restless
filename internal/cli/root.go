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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if maybeAutopilot(args) {
				return fmt.Errorf("")
			}
			return nil
		},
		Use:   "restless",
		Short: "REST API discovery and exploration CLI",
	}

	cmd.PersistentFlags().StringP("api", "a", "", "API context")
	cmd.PersistentFlags().StringP("cache", "c", "", "cache directory")

	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewEngineCmd())
	cmd.AddCommand(NewGraphCmd())

	cmd.AddCommand(NewLearnCmd())
	cmd.AddCommand(NewEngineCmd())
	cmd.AddCommand(NewGraphCmd())

	cmd.AddCommand(NewShellCmd())
	cmd.AddCommand(NewEngineCmd())
	cmd.AddCommand(NewGraphCmd())

	cmd.AddCommand(NewMapCmd())
	cmd.AddCommand(NewEngineCmd())
	cmd.AddCommand(NewGraphCmd())

	cmd.AddCommand(NewCallCmd())
	cmd.AddCommand(NewEngineCmd())
	cmd.AddCommand(NewGraphCmd())

	cmd.AddCommand(NewInspectCmd())
	cmd.AddCommand(NewEngineCmd())
	cmd.AddCommand(NewGraphCmd())

	cmd.AddCommand(NewCouncilCmd())
	cmd.AddCommand(NewEngineCmd())
	cmd.AddCommand(NewGraphCmd())

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
