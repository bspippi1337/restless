package cli

import (
	"context"
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

			fmt.Println("restless discovery engine v3")
			fmt.Println("target:", target)
			fmt.Println()

			// using queue crawler v4
			endpoints := discovery.CrawlQueueV4(target, 8)
			result := endpoints
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			start := time.Now()

			elapsed := time.Since(start)

			fmt.Println("discovered endpoints:", len(result))
			fmt.Println("scan time:", elapsed)
			fmt.Println()

			for _, e := range result {
				fmt.Println(" ", e.Path)
			}

			return nil
		},
	}

	return cmd
}
