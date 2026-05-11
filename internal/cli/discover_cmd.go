package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/discoverwow"
	"github.com/spf13/cobra"
)

func NewDiscoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover [target]",
		Short: "Semantic API discovery engine",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := discoverwow.Discover(args[0])
			if err != nil {
				return err
			}

			fmt.Print(discoverwow.Render(res))
			return nil
		},
	}

	return cmd
}
