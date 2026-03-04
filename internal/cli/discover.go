package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/core/discover"
	"github.com/spf13/cobra"
)

func NewDiscoverCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "discover <url>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			graph, err := discover.Run(args[0])
			if err != nil {
				return err
			}

			fmt.Println("Discovered endpoints:")
			for _, e := range graph.Endpoints {
				fmt.Println(" ", e)
			}

			return nil
		},
	}

	return cmd
}
