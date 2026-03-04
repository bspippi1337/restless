
package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/bspippi1337/restless/internal/core/ghactions"
	"github.com/spf13/cobra"
)

func NewSmartCmd() *cobra.Command {

	var repo string
	var workflow string
	var ref string
	var maxRequests string
	var maxDepth string

	cmd := &cobra.Command{
		Use: "smart <url>",
		Short: "Trigger remote smart probe via GitHub Actions",
		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				return fmt.Errorf("GITHUB_TOKEN not set")
			}

			target := args[0]

			client := ghactions.New(repo, token)

			inputs := map[string]string{
				"target_url": target,
				"max_requests": maxRequests,
				"max_depth": maxDepth,
			}

			err := client.Dispatch(workflow, ref, inputs)
			if err != nil {
				return err
			}

			fmt.Println("workflow dispatched")
			fmt.Println("target:", target)

			time.Sleep(2*time.Second)

			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "owner/repo")
	cmd.Flags().StringVar(&workflow, "workflow", "mosh-runner.yml", "workflow file")
	cmd.Flags().StringVar(&ref, "ref", "main", "git ref")
	cmd.Flags().StringVar(&maxRequests, "max-requests", "80", "max requests")
	cmd.Flags().StringVar(&maxDepth, "max-depth", "3", "max depth")

	return cmd
}
