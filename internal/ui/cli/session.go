package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newSessionCmd(state *State) *cobra.Command {
	return &cobra.Command{
		Use:   "session",
		Short: "Show active session state",
		RunE: func(cmd *cobra.Command, args []string) error {
			state.PrintHeader("Active session")

			if state.Session.BaseURL == "" {
				fmt.Println("No session base URL set.")
				fmt.Println("Run: restless probe <base-url>")
				return nil
			}

			fmt.Printf("Base URL:  %s\n", state.Session.BaseURL)
			fmt.Printf("Mode:      %s\n", state.Session.Mode)
			if state.Session.LastCall != "" {
				fmt.Printf("Last call: %s\n", state.Session.LastCall)
			}
			fmt.Printf("Requests:  %d\n", state.Session.RequestCount)
			if !state.Session.UpdatedAt.IsZero() {
				fmt.Printf("Updated:   %s\n", state.Session.UpdatedAt.Format(time.RFC3339))
			}
			return nil
		},
	}
}
