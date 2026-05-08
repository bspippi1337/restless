package cli

import (
	"fmt"
	"os"

	"github.com/bspippi1337/restless/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restless",
		Short: "Reactive API discovery and Unix observability runtime",
		Long: `Restless discovers, models, inspects, and explains API surfaces.

It performs bounded, same-host API discovery using safe HTTP methods,
and now also supports reactive filesystem execution workflows.

Restless is designed as a composable Unix-native runtime layer.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			maybeAutopilot(args)
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
	cmd.AddCommand(NewWatchCmd())
	cmd.AddCommand(NewVersionCmd())
	cmd.AddCommand(NewGNUCmd())

	cmd.Run = func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.Version = version.String()

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
