package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bspippi1337/restless/internal/botanical"
	"github.com/spf13/cobra"
)

// NewDiffCmd compares two previously indexed archives without API traffic.
func NewDiffCmd() *cobra.Command {
	var beforeDir, afterDir string
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare documented API operations between two offline catalogs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if beforeDir == "" || afterDir == "" {
				return fmt.Errorf("--before and --after catalog directories are required")
			}
			before, err := botanical.LoadCatalog(beforeDir)
			if err != nil {
				return fmt.Errorf("load before catalog: %w", err)
			}
			after, err := botanical.LoadCatalog(afterDir)
			if err != nil {
				return fmt.Errorf("load after catalog: %w", err)
			}
			difference, err := botanical.CompareCatalogs(before, after)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(difference)
		},
	}
	cmd.Flags().StringVar(&beforeDir, "before", "", "older catalog directory")
	cmd.Flags().StringVar(&afterDir, "after", "", "newer catalog directory")
	return cmd
}
