package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewCompletionCmd(root *cobra.Command) *cobra.Command {
	var outDir string

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate bash + zsh completion scripts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outDir == "" {
				outDir = filepath.Join("dist", "completion")
			}
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return err
			}

			r := NewRootCmd()

			bashPath := filepath.Join(outDir, "restless.bash")
			zshPath := filepath.Join(outDir, "_restless")

			if err := writeCompletion(bashPath, func(f *os.File) error { return r.GenBashCompletion(f) }); err != nil {
				return err
			}
			if err := writeCompletion(zshPath, func(f *os.File) error { return r.GenZshCompletion(f) }); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Wrote:\n  %s\n  %s\n", bashPath, zshPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&outDir, "out", "", "output directory (default: dist/completion)")
	return cmd
}

func writeCompletion(path string, gen func(*os.File) error) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gen(f)
}
