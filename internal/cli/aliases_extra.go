package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewTeachCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "teach",
		Short: "explain the latest scan result",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "teach: explain mode is available; run scan first to populate API state")
		},
	}
}

func NewCopilotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copilot",
		Short: "suggest next API exploration commands",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "copilot: try scan, then map, inspect, or fuzz")
		},
	}
}
