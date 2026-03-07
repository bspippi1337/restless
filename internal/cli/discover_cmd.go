package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/discovery"
	"github.com/bspippi1337/restless/internal/ui"
)

func NewDiscoverCmd() *cobra.Command {

	cmd := &cobra.Command{

		Use:   "discover <url>",
		Short: "reverse engineer an API",

		RunE: func(cmd *cobra.Command, args []string) error {

			target := args[0]

			ui.Banner()

			ui.Start()

			fmt.Println("target:", target)
			fmt.Println()

			start := time.Now()

			res, err := discovery.Discover(target)
			if err != nil {
				return err
			}
			_ = res.Endpoints
			if err != nil {
				return err
			}

			elapsed := time.Since(start)

			fmt.Println()
			fmt.Println()
			fmt.Println("discovered endpoints:", len(res.Endpoints))
			fmt.Println("scan time:", elapsed)

			for _, e := range res.Endpoints {

				fmt.Println(" ", e)

			}

			return nil

		},
	}

	return cmd
}
