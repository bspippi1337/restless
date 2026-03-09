package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/engine"
	"github.com/spf13/cobra"
)

func NewEngineCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "engine <target>",
		Short: "run full restless discovery engine",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			target := args[0]

			fmt.Println("RESTLESS ENGINE")
			fmt.Println("Target:", target)
			fmt.Println()

			r, err := engine.Run(target)
			if err != nil {
				return err
			}

			engine.Print(r)

			return nil
		},
	}

	return cmd
}
