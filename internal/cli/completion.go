package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewCompletionCmd(root *cobra.Command) *cobra.Command {

	return &cobra.Command{

		Use:   "completion",
		Short: "Generate bash and zsh completion",

		RunE: func(cmd *cobra.Command, args []string) error {

			dir := "dist/completion"
			os.MkdirAll(dir, 0755)

			bash := filepath.Join(dir, "restless.bash")
			zsh := filepath.Join(dir, "_restless")

			bf, _ := os.Create(bash)
			root.GenBashCompletion(bf)

			zf, _ := os.Create(zsh)
			root.GenZshCompletion(zf)

			return nil
		},
	}
}