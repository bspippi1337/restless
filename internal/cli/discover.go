package cli

import (
	"fmt"
	"sort"

	"github.com/bspippi1337/restless/internal/core/discover"
	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

func NewDiscoverCmd() *cobra.Command {
	var maxReq int
	var maxDepth int

	cmd := &cobra.Command{
		Use:   "discover <url>",
		Short: "Autonomous API discovery (crawl hypermedia + pagination + templates).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opt := discover.DefaultOptions()
			if maxReq > 0 {
				opt.MaxRequests = maxReq
			}
			if maxDepth > 0 {
				opt.MaxDepth = maxDepth
			}

			g, err := discover.RunWith(cmd.Context(), args[0], opt)
			if err != nil {
				return err
			}

			// Persist into state so `restless map` works
			st, _, _ := state.Load()
			st.LastScan.BaseURL = g.BaseURL
			st.LastScan.Endpoints = nil

			for _, p := range g.Endpoints {
				st.LastScan.Endpoints = append(st.LastScan.Endpoints, state.Route{Method: "GET", Path: p})
			}

			path, _ := state.Save(st)

			// Pretty output
			sort.Strings(g.Endpoints)
			fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\n", path)
			fmt.Fprintf(cmd.OutOrStdout(), "Visited URLs: %d\n", g.Visited)
			fmt.Fprintf(cmd.OutOrStdout(), "Discovered endpoints: %d\n", len(g.Endpoints))
			for _, n := range g.Notes {
				fmt.Fprintf(cmd.OutOrStdout(), "Note: %s\n", n)
			}
			for _, p := range g.Endpoints {
				fmt.Fprintln(cmd.OutOrStdout(), p)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&maxReq, "max-requests", 60, "max HTTP requests during discovery")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 3, "max crawl depth from seeds")
	return cmd
}
