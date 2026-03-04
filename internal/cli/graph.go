package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/core/graph"
	"github.com/spf13/cobra"
)

func NewGraphCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use: "graph",
		RunE: func(cmd *cobra.Command, args []string) error {

			out, err := graph.Render()
			if err != nil {
				return err
			}

			fmt.Println(out)

			return nil
		},
	}

	return cmd
}
