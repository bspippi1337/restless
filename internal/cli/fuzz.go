package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/core/fuzz"
	"github.com/spf13/cobra"
)

func NewFuzzCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "fuzz <url>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			res, err := fuzz.Run(args[0])
			if err != nil {
				return err
			}

			for _, r := range res {
				fmt.Printf("FOUND %s\n", r)
			}

			return nil
		},
	}

	return cmd
}
