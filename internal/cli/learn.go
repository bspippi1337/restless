package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/restlesscore"
)

func NewLearnCmd() *cobra.Command {

	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "learn <host>",
		Short: "Adaptive endpoint learner",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			r, err := restlesscore.Scan(args[0], timeout)
			if err != nil {
				return err
			}

			fmt.Fprint(
				cmd.OutOrStdout(),
				restlesscore.Render("RESTLESS LEARN", r),
			)

			return nil
		},
	}

	cmd.Flags().DurationVar(
		&timeout,
		"timeout",
		7*time.Second,
		"HTTP timeout",
	)

	return cmd
}
