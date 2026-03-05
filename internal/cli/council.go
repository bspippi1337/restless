package cli

import (
	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/app"
	"github.com/bspippi1337/restless/internal/council"
)

func NewCouncilCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "council",
		Short: "run engine council consensus meeting",
		Run: func(cmd *cobra.Command, args []string) {

			c := council.NewCouncil(app.GlobalBlackboard)
			c.Convene()

		},
	}

	return cmd
}
