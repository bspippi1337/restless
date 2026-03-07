package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/discovery"
)

func NewLearnCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "learn <url>",
		Short: "Discover API and store endpoints",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			target := args[0]

			fmt.Println("restless learning mode")
			fmt.Println("target:", target)
			fmt.Println()

			res, err := discovery.Discover(target)
			if err != nil {
				return err
			}
			_ = res.Endpoints
			if err != nil {
				return err
			}

			fmt.Println("learned endpoints:", len(res.Endpoints))
			fmt.Println()

			for _, e := range res.Endpoints {
				fmt.Println(" ", e)
			}

			return nil
		},
	}

	return cmd
}
