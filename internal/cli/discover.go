package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bspippi1337/restless/internal/core/discover"
	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

func NewDiscoverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "discover <url>",
		Short: "Autonomous API discovery (crawl JSON for same-host links).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base := strings.TrimRight(args[0], "/")
			g, err := discover.Run(base)
			if err != nil {
				return err
			}

			// Persist: overwrite last scan (avoid duplicates across runs)
			st, _, _ := state.Load()
			st.LastScan.BaseURL = g.BaseURL
			st.LastScan.Endpoints = nil
			for _, p := range g.Endpoints {
				st.LastScan.Endpoints = append(st.LastScan.Endpoints, state.Route{Method: "GET", Path: p})
			}
			_, _ = state.Save(st)

			sort.Strings(g.Endpoints)
			fmt.Fprintf(cmd.OutOrStdout(), "Visited URLs: %d\n", g.Visited)
			fmt.Fprintf(cmd.OutOrStdout(), "Discovered endpoints: %d\n", len(g.Endpoints))
			for _, e := range g.Endpoints {
				fmt.Fprintln(cmd.OutOrStdout(), e)
			}
			return nil
		},
	}
}
