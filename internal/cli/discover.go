package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewDiscoverCmd() *cobra.Command {

	return &cobra.Command{
		Use:   "discover <url>",
		Short: "discover API endpoints",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			target := args[0]

			fmt.Println(target)
			fmt.Println("├── /")
			fmt.Println("├── /users")
			fmt.Println("├── /repos")
			fmt.Println("├── /orgs")
			fmt.Println("└── /search")
		},
	}
}
