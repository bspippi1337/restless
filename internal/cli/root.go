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

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "restless",
		Short:         "Discover, map, inspect, and call APIs from the terminal",
		Long:          "Restless is an API reconnaissance CLI for exploring unknown API surfaces, building topology, and turning mystery endpoints into something usable.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringP("api", "a", "", "API context to read from cache-backed workflows")
	cmd.PersistentFlags().StringP("cache", "c", "", "cache directory")

	core := []*cobra.Command{
		NewDiscoverCmd(),
		NewLearnCmd(),
		NewMapCmd(),
		NewInspectCmd(),
		NewCallCmd(),
		NewShellCmd(),
		NewCompletionCmd(cmd),
	}

	experimental := []*cobra.Command{
		NewBlckswanCmd(),
		NewSmartCmd(),
		NewEngineCmd(),
		NewGraphCmd(),
		NewCouncilCmd(),
	}

	for _, sub := range core {
		cmd.AddCommand(sub)
	}
	for _, sub := range experimental {
		cmd.AddCommand(sub)
	}

	cmd.SetVersionTemplate("restless {{.Version}}\n")
	cmd.Version = fmt.Sprintf("%s (%s %s)", version, commit, date)

	AddDynamicCommands(cmd)
	return cmd
}

func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
