package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

func NewMapCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use: "map",
		RunE: func(cmd *cobra.Command, args []string) error {

			st, _, err := state.Load()
			if err != nil {
				return err
			}

			for _, r := range st.LastScan.Endpoints {
				fmt.Printf("%s %s\n", r.Method, r.Path)
			}

			return nil
		},
	}

	return cmd
}
