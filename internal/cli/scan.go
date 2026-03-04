package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/core/scan"
	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

func NewScanCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "scan <url>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			res, err := scan.Run(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			st, _, _ := state.Load()
			st.LastScan = res
			path, _ := state.Save(st)

			fmt.Printf("Saved scan → %s\n", path)
			fmt.Printf("Routes discovered: %d\n", len(res.Endpoints))

			return nil
		},
	}

	return cmd
}
