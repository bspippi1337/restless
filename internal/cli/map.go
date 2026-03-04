package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

// map prints an endpoint list from the last scan/discover state.
// Output is stable (sorted) and grouped by top-level prefix.
func NewMapCmd() *cobra.Command {
	var tree bool

	cmd := &cobra.Command{
		Use:   "map",
		Short: "Print endpoint map from last discover/scan.",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, _, err := state.Load()
			if err != nil {
				return err
			}
			if st.LastScan.BaseURL == "" || len(st.LastScan.Endpoints) == 0 {
				return fmt.Errorf("no endpoints in state. Run: restless discover <url>")
			}

			paths := make([]string, 0, len(st.LastScan.Endpoints))
			for _, r := range st.LastScan.Endpoints {
				if strings.TrimSpace(r.Path) != "" {
					paths = append(paths, r.Path)
				}
			}
			sort.Strings(paths)

			fmt.Fprintf(cmd.OutOrStdout(), "Base: %s\n", st.LastScan.BaseURL)
			if !tree {
				for _, p := range paths {
					fmt.Fprintln(cmd.OutOrStdout(), p)
				}
				return nil
			}

			// Tree-ish grouping by first segment
			groups := map[string][]string{}
			for _, p := range paths {
				seg := strings.Split(strings.TrimPrefix(p, "/"), "/")[0]
				if seg == "" {
					seg = "/"
				}
				groups[seg] = append(groups[seg], p)
			}

			keys := make([]string, 0, len(groups))
			for k := range groups {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "\n[%s]\n", k)
				for _, p := range groups[k] {
					fmt.Fprintln(cmd.OutOrStdout(), " ", p)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&tree, "tree", true, "group by top-level prefix")
	return cmd
}
