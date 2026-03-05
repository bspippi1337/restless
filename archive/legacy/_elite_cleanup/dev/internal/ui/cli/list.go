package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd(state *State) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List discovered endpoints for the active session",
		RunE: func(cmd *cobra.Command, args []string) error {
			state.PrintHeader("Surface")
			if state.Session.BaseURL == "" {
				fmt.Println("No base URL in session. Run: restless probe <base-url>")
				return nil
			}
			fmt.Printf("Base: %s\n\n", state.Session.BaseURL)
			printEndpointTable(state.Session.Endpoints)
			fmt.Println()
			fmt.Println("Tip: run `restless run GET /path`")
			return nil
		},
	}
}

func printEndpointTable(eps []Endpoint) {
	if len(eps) == 0 {
		fmt.Println("(no endpoints discovered yet)")
		return
	}
	fmt.Printf("%-8s %s\n", "METHOD", "PATH")
	fmt.Printf("%-8s %s\n", "------", "------------------")
	for _, ep := range eps {
		fmt.Printf("%-8s %s\n", ep.Method, ep.Path)
	}
}
