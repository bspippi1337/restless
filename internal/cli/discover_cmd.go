package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/discovery"
)

func NewDiscoverCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "discover <url>",
		Short: "reverse-engineer and map an unknown API",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			target := args[0]

			fmt.Println("restless discovery engine v4")
			fmt.Println("target:", target)
			fmt.Println()

			start := time.Now()

			endpoints := discovery.CrawlQueueV4(target, 8)

			elapsed := time.Since(start)

			fmt.Println()
			fmt.Println("discovered endpoints:", len(endpoints))
			fmt.Println("scan time:", elapsed)
			fmt.Println()

			for _, e := range endpoints {
				fmt.Println(" ", e.Path)
			}

			return nil
		},
	}

	return cmd
}
