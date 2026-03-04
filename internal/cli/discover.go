package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/core/discover"
	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

func NewDiscoverCmd() *cobra.Command {

	return &cobra.Command{

		Use:  "discover <url>",
		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			g, err := discover.Run(args[0])
			if err != nil {
				return err
			}

			st, _, _ := state.Load()

			for _, p := range g.Endpoints {
				st.LastScan.Endpoints = append(st.LastScan.Endpoints, state.Route{
					Method: "GET",
					Path:   p,
				})
			}

			state.Save(st)

			fmt.Println("Visited URLs:", g.Visited)
			fmt.Println("Discovered endpoints:")

			for _, e := range g.Endpoints {
				fmt.Println(" ", e)
			}

			return nil
		},
	}
}