package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("restless %s\ncommit: %s\ndate: %s\n", buildVersion, buildCommit, buildDate)
		},
	}
}

func NewGNUCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gnu",
		Short: "Show GNU-style info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("restless: unix-style file watcher and automation engine")
		},
	}
}
