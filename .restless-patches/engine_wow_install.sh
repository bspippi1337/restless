#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

cat > internal/enginewow/engine.go <<'GO'
package enginewow

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Node struct {
	Path        string
	Status      int
	Latency     time.Duration
	Bytes       int
	ContentType string
	Class       string
	Signal      string
	Confidence  int
	AuthWall    bool
	Alive       bool
}

type Result struct {
	Target      string
	Base        string
	Kind        string
	Traits      []string
	Nodes       []Node
	AvgLatency  time.Duration
	AuthScore   int
	Surface     int
	Exposure    string
	Suggestions []string
}

var seeds = []string{
	"/",
	"/api",
	"/api/v1",
	"/api/v2",
	"/v1",
	"/v2",
	"/health",
	"/status",
	"/version",
	"/openapi.json",
	"/swagger.json",
	"/docs",
	"/users",
	"/user",
	"/repos",
	"/projects",
	"/search",
	"/events",
	"/rate_limit",
	"/auth",
	"/login",
	"/me",
}

func Crawl(ctx context.Context, raw string, timeout time.Duration) (*Result, error) {
	base, err := normalize(raw)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}

	r := &Result{
		Target: raw,
		Base:   base,
		Kind:   "Unknown",
	}

	for _, p := range seeds {
		n := probe(ctx, client, base, p)
		if n.Status > 0 {
			n.Class, n.Signal, n.Confidence = classify(n)
			n.AuthWall = n.Status == 401 || n.Status == 403
			n.Alive = n.Status < 500
			r.Nodes = append(r.Nodes, n)
		}
	}

	sort.Slice(r.Nodes, func(i, j int) bool {
		return r.Nodes[i].Confidence > r.Nodes[j].Confidence
	})

	analyze(r)
	return r, nil
}

func probe(ctx context.Context, client *http.Client, base, path string) Node {
	start := time.Now()

	req, _ := http.NewRequestWithContext(ctx, "GET", base+path, nil)
	req.Header.Set("User-Agent", "restless-engine/next")

	resp, err := client.Do(req)
	if err != nil {
		return Node{Path: path}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	return Node{
		Path:        path,
		Status:      resp.StatusCode,
		Latency:     time.Since(start),
		Bytes:       len(body),
		ContentType: resp.Header.Get("Content-Type"),
	}
}

func classify(n Node) (string, string, int) {
	p := strings.ToLower(n.Path)

	switch {
	case strings.Contains(p, "openapi"), strings.Contains(p, "swagger"):
		return "schema", "contract found", 96
	case strings.Contains(p, "health"):
		return "ops", "service pulse", 88
	case strings.Contains(p, "user"):
		return "identity", "user namespace", 86
	case strings.Contains(p, "repo"):
		return "resource", "repository domain", 84
	case strings.Contains(p, "search"):
		return "index", "search backend", 82
	case strings.Contains(p, "event"):
		return "stream", "activity feed", 78
	}

	return "endpoint", "candidate", 60
}

func analyze(r *Result) {
	var total time.Duration
	var auth int

	for _, n := range r.Nodes {
		total += n.Latency
		if n.AuthWall {
			auth++
		}
	}

	if len(r.Nodes) > 0 {
		r.AvgLatency = total / time.Duration(len(r.Nodes))
		r.AuthScore = int(math.Round(float64(auth) / float64(len(r.Nodes)) * 100))
	}

	r.Surface = len(r.Nodes)
	r.Kind = "REST/JSON"

	switch {
	case r.Surface >= 10:
		r.Exposure = "wide"
	case r.Surface >= 5:
		r.Exposure = "medium"
	default:
		r.Exposure = "small"
	}

	if r.AuthScore > 50 {
		r.Suggestions = append(r.Suggestions,
			"Authenticated crawl may dramatically expand the map.")
	}
}

func Render(r *Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\nRESTLESS ENGINE\n")
	fmt.Fprintf(&b, "Target     %s\n", r.Base)
	fmt.Fprintf(&b, "Type       %s\n", r.Kind)
	fmt.Fprintf(&b, "Surface    %d live nodes\n", r.Surface)
	fmt.Fprintf(&b, "Exposure   %s\n", r.Exposure)
	fmt.Fprintf(&b, "Auth wall  %d%%\n", r.AuthScore)
	fmt.Fprintf(&b, "Latency    %s avg\n\n", r.AvgLatency.Round(time.Millisecond))

	fmt.Fprintf(&b, "Recon\n─────\n")

	for _, n := range r.Nodes {
		fmt.Fprintf(&b,
			"  %s %-16s %-3d %-10s %s\n",
			icon(n),
			n.Path,
			n.Status,
			n.Class,
			n.Signal,
		)
	}

	fmt.Fprintf(&b, "\nSurface Map\n───────────\n")
	fmt.Fprintf(&b, "  ◎ /\n")

	for _, n := range r.Nodes {
		if n.Path == "/" {
			continue
		}

		fmt.Fprintf(&b,
			"  ├─%s %-18s %s\n",
			dot(n),
			n.Path,
			bar(n.Confidence),
		)
	}

	fmt.Fprintf(&b, "\nVerdict\n───────\n")

	if r.AuthScore > 50 {
		fmt.Fprintf(&b,
			"  API is alive, structured, and guarded.\n")
	} else {
		fmt.Fprintf(&b,
			"  Good visible surface for adaptive traversal.\n")
	}

	return b.String()
}

func normalize(raw string) (string, error) {
	if !strings.HasPrefix(raw, "http") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	u.Path = ""

	return strings.TrimRight(u.String(), "/"), nil
}

func icon(n Node) string {
	if n.Status == 200 {
		return "✓"
	}
	if n.Status == 401 || n.Status == 403 {
		return "⛔"
	}
	return "•"
}

func dot(n Node) string {
	if n.AuthWall {
		return "◉"
	}
	return "◆"
}

func bar(v int) string {
	filled := v / 10

	if filled < 1 {
		filled = 1
	}

	if filled > 10 {
		filled = 10
	}

	return strings.Repeat("█", filled) +
		strings.Repeat("░", 10-filled)
}
GO

cat > cmd/restless/engine_wow.go <<'GO'
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bspippi1337/restless/internal/enginewow"
	"github.com/spf13/cobra"
)

var engineCmd = &cobra.Command{
	Use:   "engine [target]",
	Short: "Adaptive API reconnaissance engine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("timeout")

		res, err := enginewow.Crawl(
			context.Background(),
			args[0],
			timeout,
		)

		if err != nil {
			return err
		}

		fmt.Print(enginewow.Render(res))
		return nil
	},
}

func init() {
	engineCmd.Flags().Duration(
		"timeout",
		7*time.Second,
		"request timeout",
	)

	rootCmd.AddCommand(engineCmd)
}
GO

gofmt -w internal/enginewow/engine.go
gofmt -w cmd/restless/engine_wow.go

echo
echo "[+] Installed enginewow"
echo
echo "Build:"
echo "  go build -o build/restless ./cmd/restless"
echo
echo "Test:"
echo "  ./build/restless engine api.github.com"
