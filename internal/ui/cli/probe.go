package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	httpadapter "github.com/bspippi1337/restless/internal/adapters/http"
	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/spf13/cobra"
)

var commonSpecPaths = []string{
	"/openapi.json",
	"/openapi.yaml",
	"/swagger.json",
	"/swagger.yaml",
	"/v3/api-docs",
	"/api-docs",
}

var heuristicCandidates = []Endpoint{
	{Method: "GET", Path: "/health", Source: "heuristic"},
	{Method: "GET", Path: "/status", Source: "heuristic"},
	{Method: "GET", Path: "/metrics", Source: "heuristic"},
	// httpbin-friendly (great for demos)
	{Method: "GET", Path: "/get", Source: "heuristic"},
	{Method: "POST", Path: "/post", Source: "heuristic"},
	{Method: "GET", Path: "/headers", Source: "heuristic"},
	{Method: "GET", Path: "/ip", Source: "heuristic"},
	{Method: "GET", Path: "/user-agent", Source: "heuristic"},
	{Method: "GET", Path: "/status/200", Source: "heuristic"},
}

func newProbeCmd(state *State) *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "probe [base-url]",
		Short: "Probe a base URL (spec detection + heuristic discovery)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base := strings.TrimRight(args[0], "/")
			if _, err := url.ParseRequestURI(base); err != nil {
				return fmt.Errorf("invalid base url: %w", err)
			}

			state.PrintHeader(fmt.Sprintf("Probing %s", base))
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Lightweight connectivity check
			if err := ping(ctx, base); err != nil {
				return err
			}
			fmt.Println("✓ Connected")

			// Spec detection
			specURL, ok := detectSpec(ctx, base)
			if ok {
				fmt.Printf("✓ Spec detected: %s\n", specURL)
				state.Session.BaseURL = base
				state.Session.Mode = ModeSpec

				// Minimal spec extraction (paths only if JSON)
				eps, err := extractSpecEndpoints(ctx, specURL)
				if err == nil && len(eps) > 0 {
					state.Session.Endpoints = eps
					fmt.Printf("✓ Endpoints: %d\n\n", len(eps))
					printEndpointsPreview(eps, 8)
				} else {
					fmt.Println("✓ Spec reachable (endpoint extraction pending)")
					state.Session.Endpoints = nil
				}
				fmt.Println("\nSession initialized.")
				return state.Save()
			}

			fmt.Println("✓ No OpenAPI detected")
			state.Session.BaseURL = base
			state.Session.Mode = ModeHeuristic

			// Heuristic discovery: test a small set of candidates
			eps := discoverHeuristic(ctx, base)
			state.Session.Endpoints = eps
			fmt.Printf("✓ Discovered endpoints: %d\n\n", len(eps))
			printEndpointsPreview(eps, 8)
			fmt.Println("\nSession initialized.")
			return state.Save()
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 6*time.Second, "Probe timeout")
	return cmd
}

func ping(ctx context.Context, base string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", base+"/", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}

func detectSpec(ctx context.Context, base string) (string, bool) {
	for _, p := range commonSpecPaths {
		u := base + p
		req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			continue
		}
		ct := resp.Header.Get("Content-Type")
		isJSON := strings.Contains(ct, "json") || strings.HasPrefix(strings.TrimSpace(string(b)), "{")
		if !isJSON {
			continue
		}
		// quick fingerprint
		s := string(b)
		if strings.Contains(s, ""openapi"") || strings.Contains(s, ""swagger"") {
			return u, true
		}
	}
	return "", false
}

// extractSpecEndpoints is intentionally minimal for v0.2.
// If it fails, we still keep ModeSpec and let the user run by full URL.
func extractSpecEndpoints(ctx context.Context, specURL string) ([]Endpoint, error) {
	// Use engine + transport so we stay in-project.
	transport := &httpadapter.HTTPTransport{}
	eng := &engine.Engine{Transport: transport}
	res := eng.Run(ctx, engine.Job{Method: "GET", Target: specURL})
	if res.Err != nil {
		return nil, res.Err
	}
	b := strings.TrimSpace(string(res.Body))
	if !strings.HasPrefix(b, "{") {
		return nil, fmt.Errorf("spec not json")
	}

	// Super-light JSON scanning for "/paths" keys would be brittle without a real parser.
	// We do a simple heuristic: collect strings that look like "\/something" under "paths".
	// If it fails, return empty list without error.
	paths := make(map[string]bool)

	// Find the "paths" object region (best-effort).
	idx := strings.Index(b, ""paths"")
	if idx < 0 {
		return nil, nil
	}
	sub := b[idx:]
	// collect tokens like "/foo" within a limited window
	limit := sub
	if len(limit) > 120000 {
		limit = limit[:120000]
	}
	for _, tok := range strings.Split(limit, """) {
		if strings.HasPrefix(tok, "/") && len(tok) <= 200 && !strings.Contains(tok, " ") {
			paths[tok] = true
		}
	}
	if len(paths) == 0 {
		return nil, nil
	}

	out := make([]Endpoint, 0, len(paths))
	for p := range paths {
		out = append(out, Endpoint{Method: "*", Path: p, Source: "spec"})
	}
	// keep deterministic-ish order
	sortEndpoints(out)
	return out, nil
}

func discoverHeuristic(ctx context.Context, base string) []Endpoint {
	transport := &httpadapter.HTTPTransport{}
	eng := &engine.Engine{Transport: transport}

	seen := map[string]bool{}
	out := make([]Endpoint, 0, len(heuristicCandidates))

	for _, ep := range heuristicCandidates {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			continue
		}
		seen[key] = true

		target := base + ep.Path
		res := eng.Run(ctx, engine.Job{Method: ep.Method, Target: target})
		if res.Err != nil {
			continue
		}
		if res.Status >= 200 && res.Status < 400 {
			out = append(out, ep)
		}
	}
	sortEndpoints(out)
	return out
}

func printEndpointsPreview(eps []Endpoint, max int) {
	if len(eps) == 0 {
		fmt.Println("(none found yet)")
		return
	}
	if max > len(eps) {
		max = len(eps)
	}
	for i := 0; i < max; i++ {
		ep := eps[i]
		fmt.Printf("  %-6s %s\n", ep.Method, ep.Path)
	}
	if len(eps) > max {
		fmt.Printf("  …and %d more\n", len(eps)-max)
	}
}
