package cli

import (
	"fmt"
	"strings"

	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

func NewInspectCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "inspect <METHOD> <PATH>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			method := strings.ToUpper(args[0])
			path := args[1]

			st, _, err := state.Load()
			if err != nil {
				return err
			}

			for _, r := range st.LastScan.Endpoints {
				if r.Method == method && r.Path == path {

					fmt.Println("Route found")
					fmt.Println("Example:")
					fmt.Printf("curl %s%s\n", st.LastScan.BaseURL, path)

					return nil
				}
			}

			return fmt.Errorf("route not found")
		},
	}

	return cmd
}
