package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewAutoCmd() *cobra.Command {

	var repo string
	var matrix bool

	cmd := &cobra.Command{
		Use:   "auto <url>",
		Short: "Autonomous recon: discover endpoints then launch swarm probing",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			target := args[0]

			fmt.Println("⚡ restless autonomous mode")
			fmt.Println("target:", target)

			// naive endpoint discovery seeds
			seeds := []string{
				target + "/api",
				target + "/v1",
				target + "/v2",
				target + "/graphql",
			}

			fmt.Println("generated targets:")
			for _, s := range seeds {
				fmt.Println(" -", s)
			}

			if repo == "" {
				fmt.Println("tip: add --repo owner/repo to dispatch swarm")
				return nil
			}

			args2 := []string{"swarm"}
			args2 = append(args2, seeds...)
			args2 = append(args2, "--repo", repo)

			if matrix {
				args2 = append(args2, "--matrix")
			}

			fmt.Println("\nlaunching swarm...")
			root := NewRootCmd()
			root.SetArgs(args2)
			return root.Execute()
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "GitHub repo")
	cmd.Flags().BoolVar(&matrix, "matrix", true, "use matrix swarm")

	return cmd
}
