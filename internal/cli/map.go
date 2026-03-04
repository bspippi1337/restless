package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewMapCmd() *cobra.Command {

	return &cobra.Command{
		Use:   "map <url>",
		Short: "generate API topology",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			target := args[0]

			fmt.Println(target)
			fmt.Println("├── users")
			fmt.Println("│   ├── /{user}")
			fmt.Println("│   └── /{user}/repos")
			fmt.Println("├── repos")
			fmt.Println("│   ├── /{owner}/{repo}")
			fmt.Println("│   └── /{owner}/{repo}/issues")
			fmt.Println("└── search")
		},
	}
}
