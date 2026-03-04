
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bspippi1337/restless/internal/core/ghactions"
	"github.com/spf13/cobra"
)

func NewSwarmCmd() *cobra.Command {

	var repo string
	var workflow string
	var ref string
	var matrix bool

	cmd := &cobra.Command{
		Use: "swarm <url> [url...]",
		Short: "Distributed API probing swarm",
		Args: cobra.MinimumNArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				return fmt.Errorf("GITHUB_TOKEN not set")
			}

			client := ghactions.New(repo, token)

			if matrix {

				matrixJSON,_ := json.Marshal(args)

				inputs := map[string]string{
					"targets_json": string(matrixJSON),
				}

				err := client.Dispatch(workflow, ref, inputs)
				if err != nil {
					return err
				}

				fmt.Println("matrix swarm dispatched")
				fmt.Println("targets:", strings.Join(args,", "))

				return nil
			}

			for _,t := range args {

				inputs := map[string]string{
					"target_url": t,
				}

				err := client.Dispatch(workflow, ref, inputs)
				if err != nil {
					fmt.Println("dispatch failed:",t)
					continue
				}

				fmt.Println("dispatched:",t)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&repo,"repo","", "owner/repo")
	cmd.Flags().StringVar(&workflow,"workflow","mosh-runner.yml","workflow file")
	cmd.Flags().StringVar(&ref,"ref","main","git ref")
	cmd.Flags().BoolVar(&matrix,"matrix",false,"dispatch as GitHub Actions matrix job")

	return cmd
}
