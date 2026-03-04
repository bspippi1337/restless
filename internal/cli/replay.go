package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewReplayCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "replay <curl>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			out, err := exec.Command("sh", "-c", args[0]).CombinedOutput()
			if err != nil {
				return err
			}

			fmt.Println(string(out))

			return nil
		},
	}

	return cmd
}
