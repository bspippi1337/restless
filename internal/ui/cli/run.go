package cli

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	httpadapter "github.com/bspippi1337/restless/internal/adapters/http"
	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/spf13/cobra"
)

func newRunCmd(state *State) *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "run [method] [path-or-url]",
		Short: "Run a request (uses session base when available)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			target := args[1]

			finalURL, err := resolveTarget(state.Session.BaseURL, target)
			if err != nil {
				return err
			}

			state.PrintHeader("")
			fmt.Printf("%s %s\n", method, finalURL)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			transport := &httpadapter.HTTPTransport{}
			eng := &engine.Engine{Transport: transport}

			start := time.Now()
			res := eng.Run(ctx, engine.Job{Method: method, Target: finalURL})
			dur := time.Since(start)

			if res.Err != nil {
				return res.Err
			}

			fmt.Printf("\nStatus: %d\n", res.Status)
			fmt.Printf("Time:   %s\n", formatDuration(dur))
			fmt.Println(strings.Repeat("-", 40))
			fmt.Println()

			// Print body as-is (v0.2 polish: truncate later if needed)
			fmt.Println(string(res.Body))

			state.Session.RequestCount++
			state.Session.LastCall = fmt.Sprintf("%s %s", method, target)
			_ = state.Save()
			return nil
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Request timeout")
	return cmd
}

func resolveTarget(base, pathOrURL string) (string, error) {
	// full URL?
	if u, err := url.ParseRequestURI(pathOrURL); err == nil && u.Scheme != "" && u.Host != "" {
		return strings.TrimRight(pathOrURL, "/"), nil
	}
	if base == "" {
		return "", fmt.Errorf("no base URL in session; run `restless probe <base-url>` or pass a full URL")
	}
	p := pathOrURL
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return strings.TrimRight(base, "/") + p, nil
}
