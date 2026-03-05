package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	state := NewState()

	cmd := &cobra.Command{
		Use:           "restless",
		Short:         "Terminal-first API workbench",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return state.Load()
		},
	}

	cmd.PersistentFlags().StringVar(&state.SessionName, "session", "default", "Session name")
	cmd.PersistentFlags().BoolVar(&state.NoHeader, "no-header", false, "Disable header output")

	cmd.AddCommand(newProbeCmd(state))
	cmd.AddCommand(newListCmd(state))
	cmd.AddCommand(newRunCmd(state))
	cmd.AddCommand(newSessionCmd(state))

	cmd.SetHelpTemplate(helpTemplate())

	return cmd
}

func helpTemplate() string {
	return fmt.Sprintf(`{{with or .Long .Short }}{{. | trimTrailingWhitespaces}}{{end}}

Usage:
  {{.UseLine}}

Commands:
{{range .Commands}}{{if (and .IsAvailableCommand (not .IsHelpCommand))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}

Flags:
{{.Flags.FlagUsages | trimTrailingWhitespaces}}

`)
}
