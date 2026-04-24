package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan <target>",
		Short: "scan API surface quickly",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("scan: delegated to engine discovery pipeline")
			NewEngineCmd().Run(cmd, args)
		},
	}
}

func NewTeachCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "teach",
		Short: "explain detected API model",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("teach: interactive explanation layer (placeholder active)")
		},
	}
}

func NewFuzzCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fuzz",
		Short: "fuzz API endpoints heuristically",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("fuzz: heuristic probing layer active")
		},
	}
}

func NewCopilotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copilot",
		Short: "generate next-step API commands",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("copilot: suggestion engine active")
		},
	}
}
