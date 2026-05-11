package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type OctoProbe struct {
	Path        string            `json:"path"`
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Status      int               `json:"status"`
	Bytes       int               `json:"bytes"`
	DurationMS  int64             `json:"duration_ms"`
	ContentType string            `json:"content_type,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Sniff       string            `json:"sniff,omitempty"`
	Error       string            `json:"error,omitempty"`
}

type OctoBrain struct {
	Target      string            `json:"target"`
	Host        string            `json:"host"`
	GeneratedAt string            `json:"generated_at"`
	Arms        int               `json:"arms"`
	Seeds       int               `json:"seeds"`
	FoundPaths  []string          `json:"found_paths"`
	Signals     map[string]any    `json:"signals"`
	RateHints   map[string]string `json:"rate_hints,omitempty"`
	Probes      []OctoProbe       `json:"probes"`
}

var defaultOctoSeeds = []string{
	"/", "/api", "/v1", "/v2", "/v3",
	"/health", "/status", "/version", "/metrics",
	"/swagger.json", "/openapi.json", "/swagger/v1/swagger.json", "/v3/api-docs",
	"/graphql",
	"/users", "/repos", "/orgs", "/search",
	"/admin", "/internal",
}

func NewOctoSwanCmd() *cobra.Command {
	var outDir string
	var arms int
	var max int
	var timeout time.Duration
	var header []string
	var demo bool

	cmd := &cobra.Command{
		Use:     "octoswan <url>",
		Aliases: []string{"octopus", "octobrain", "swanbrain", "brain"},
		Short:   "Swan brain + octopus: parallel probing + signal inference + README-ready artifacts",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, baseURL, err := normalizeTarget(args[0])
			if err != nil {
				return err
			}
			if arms <= 0 {
				arms = 8
			}
			if arms > 64 {
				arms = 64
			}
			if timeout <= 0 {
				timeout = 6 * time.Second
			}

			seeds := append([]string{}, defaultOctoSeeds...)
			if max > 0 && max < len(seeds) {
				seeds = seeds[:max]
			}

			cli := &http.Client{Timeout: timeout}
			hdrs := parseHeaders(header)

			ctx := context.Background()
			probes := runOctopus(ctx, cli, baseURL, hdrs, arms, seeds)
			found := collectFound(probes)
			signals, rate := inferSignals(probes)

			rep := OctoBrain{
				Target:      target,
				Host:        baseURL.Host,
				GeneratedAt: time.Now().UTC().Format(time.RFC3339),
				Arms:        arms,
				Seeds:       len(seeds),
				FoundPaths:  found,
				Signals:     signals,
				RateHints:   rate,
				Probes:      probes,
			}

			if outDir == "" {
				outDir = "dist"
			}
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return err
			}

			stamp := time.Now().UTC().Format("20060102_150405")
			base := fmt.Sprintf("octoswan_%s_%s", safeName(baseURL.Host), stamp)

			jsonPath := filepath.Join(outDir, base+".json")
			mdPath := filepath.Join(outDir, base+".summary.md")
			svgPath := filepath.Join(outDir, base+".map.svg")

			j, _ := json.MarshalIndent(rep, "", "  ")
			_ = os.WriteFile(jsonPath, j, 0o644)
			_ = os.WriteFile(mdPath, []byte(buildSummary(rep, filepath.Base(svgPath))), 0o644)
			_ = os.WriteFile(svgPath, []byte(buildMapSVG(rep.Host, rep.FoundPaths)), 0o644)

			fmt.Println("swan brain online + octopus arms deployed")
			fmt.Println("target:", target)
			fmt.Println("arms:", arms, "seeds:", len(seeds), "found:", len(found))
			if demo {
				fmt.Println("mode: demo")
			}
			fmt.Println()

			fmt.Println("signals:")
			keys := make([]string, 0, len(signals))
			for k := range signals {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("  - %s: %v\n", k, signals[k])
			}

			if len(rate) > 0 {
				fmt.Println()
				fmt.Println("rate hints:")
				rk := make([]string, 0, len(rate))
				for k := range rate {
					rk = append(rk, k)
				}
				sort.Strings(rk)
				for _, k := range rk {
					fmt.Printf("  - %s: %s\n", k, rate[k])
				}
			}

			fmt.Println()
			fmt.Println("artifacts:")
			fmt.Println("  report: ", jsonPath)
			fmt.Println("  summary:", mdPath)
			fmt.Println("  map:    ", svgPath)

			if demo {
				fmt.Println()
				fmt.Println("demo snippet:")
				fmt.Println("  restless brain " + target)
				fmt.Println("  restless octopus --arms 16 " + target)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outDir, "out", "dist", "output directory")
	cmd.Flags().IntVar(&arms, "arms", 8, "number of parallel workers")
	cmd.Flags().IntVar(&max, "max", 0, "max seed paths (0 = all)")
	cmd.Flags().DurationVar(&timeout, "timeout", 6*time.Second, "request timeout")
	cmd.Flags().StringArrayVar(&header, "header", nil, "extra header (repeatable), e.g. --header 'Authorization: Bearer ...'")
	cmd.Flags().BoolVar(&demo, "demo", false, "print extra demo snippet for README/HN")

	return cmd
}

func normalizeTarget(raw string) (string, *url.URL, error) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return "", nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		return "", nil, fmt.Errorf("target missing host: %q", raw)
	}
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String(), u, nil
}

func parseHeaders(hs []string) map[string]string {
	m := map[string]string{}
	for _, h := range hs {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if k != "" && v != "" {
			m[k] = v
		}
	}
	return m
}

func runOctopus(ctx context.Context, cli *http.Client, base *url.URL, headers map[string]string, arms int, paths []string) []OctoProbe {
	type job struct{ path string }
	jobs := make(chan job, len(paths))
	out := make(chan OctoProbe, len(paths))

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for j := range jobs {
			u := *base
			u.Path = j.path
			req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
			if err != nil {
				out <- OctoProbe{Path: j.path, URL: u.String(), Method: "GET", Error: err.Error()}
				continue
			}
			req.Header.Set("User-Agent", "restless-octoswan/1")
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			start := time.Now()
			resp, err := cli.Do(req)
			if err != nil {
				out <- OctoProbe{Path: j.path, URL: u.String(), Method: "GET", Error: err.Error()}
				continue
			}
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
			_ = resp.Body.Close()

			out <- OctoProbe{
				Path:        j.path,
				URL:         u.String(),
				Method:      "GET",
				Status:      resp.StatusCode,
				Bytes:       len(body),
				DurationMS:  time.Since(start).Milliseconds(),
				ContentType: resp.Header.Get("Content-Type"),
				Headers:     pickHeaders(resp.Header),
				Sniff:       sniff(body),
			}
		}
	}

	wg.Add(arms)
	for i := 0; i < arms; i++ {
		go worker()
	}

	for _, p := range paths {
		jobs <- job{path: p}
	}
	close(jobs)

	wg.Wait()
	close(out)

	res := make([]OctoProbe, 0, len(paths))
	for p := range out {
		res = append(res, p)
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Status == res[j].Status {
			return res[i].Path < res[j].Path
		}
		return res[i].Status < res[j].Status
	})
	return res
}

func pickHeaders(h http.Header) map[string]string {
	out := map[string]string{}
	for _, k := range []string{
		"Server", "X-Powered-By", "WWW-Authenticate",
		"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After",
	} {
		if v := h.Get(k); v != "" {
			out[k] = v
		}
	}
	return out
}

func sniff(body []byte) string {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return ""
	}
	if len(body) > 120 {
		body = body[:120]
	}
	s := strings.ReplaceAll(string(body), "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

func collectFound(probes []OctoProbe) []string {
	found := []string{}
	for _, p := range probes {
		if p.Error != "" {
			continue
		}
		if p.Status > 0 && p.Status < 500 {
			found = append(found, p.Path)
		}
	}
	uniq := map[string]bool{}
	out := []string{}
	for _, p := range found {
		if !uniq[p] {
			uniq[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func inferSignals(probes []OctoProbe) (map[string]any, map[string]string) {
	signals := map[string]any{}
	rate := map[string]string{}

	var authHits, okHits, openapiHits, graphqlHits int
	versionsSet := map[string]bool{}
	serverSet := map[string]bool{}

	reVersion := regexp.MustCompile(`^/v([0-9]{1,2})(/|$)`)

	for _, p := range probes {
		for hk, hv := range p.Headers {
			if strings.HasPrefix(hk, "X-RateLimit") || hk == "Retry-After" {
				rate[hk] = hv
			}
			if hk == "Server" && hv != "" {
				serverSet[hv] = true
			}
		}

		if p.Error != "" {
			continue
		}
		if p.Status >= 200 && p.Status < 400 {
			okHits++
		}
		if p.Status == 401 || p.Status == 403 {
			authHits++
		}
		if strings.Contains(strings.ToLower(p.ContentType), "json") {
			low := strings.ToLower(p.Sniff)
			if strings.Contains(low, `"openapi"`) || strings.Contains(low, `"swagger"`) || strings.Contains(low, `"paths"`) {
				openapiHits++
			}
			if (p.Path == "/graphql" || strings.Contains(p.Path, "graphql")) && strings.Contains(low, `"errors"`) {
				graphqlHits++
			}
		}
		if m := reVersion.FindStringSubmatch(p.Path); len(m) == 2 {
			versionsSet["v"+m[1]] = true
		}
	}

	signals["reachable"] = okHits > 0
	signals["auth_likely"] = authHits > 0
	signals["openapi_hint"] = openapiHits > 0
	signals["graphql_hint"] = graphqlHits > 0

	if len(serverSet) > 0 {
		servers := make([]string, 0, len(serverSet))
		for s := range serverSet {
			servers = append(servers, s)
		}
		sort.Strings(servers)
		signals["server"] = servers
	}

	if len(versionsSet) > 0 {
		vs := make([]string, 0, len(versionsSet))
		for v := range versionsSet {
			vs = append(vs, v)
		}
		sort.Strings(vs)
		signals["versioning"] = vs
	} else {
		signals["versioning"] = "unknown"
	}

	return signals, rate
}

func buildSummary(rep OctoBrain, svgFile string) string {
	var b strings.Builder
	b.WriteString("# ⚡ OctoSwan Recon Summary\n\n")
	b.WriteString("Target: `" + rep.Target + "`\n\n")
	b.WriteString("Generated: `" + rep.GeneratedAt + "`\n\n")
	b.WriteString("Arms: `" + itoa(rep.Arms) + "`\n\n")

	b.WriteString("## Map\n\n")
	b.WriteString("```html\n")
	b.WriteString("<p align=\"center\"><img src=\"" + svgFile + "\" alt=\"OctoSwan API map\"></p>\n")
	b.WriteString("```\n\n")

	b.WriteString("## Signals\n\n")
	keys := make([]string, 0, len(rep.Signals))
	for k := range rep.Signals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("- **" + k + "**: " + fmtAny(rep.Signals[k]) + "\n")
	}
	b.WriteString("\n")

	b.WriteString("## Found paths\n\n")
	for _, p := range rep.FoundPaths {
		b.WriteString("- " + p + "\n")
	}
	b.WriteString("\n")

	b.WriteString("## Next moves\n\n")
	b.WriteString("- Add auth headers if you see 401/403.\n")
	b.WriteString("- Increase arms for speed: `--arms 16`.\n")
	b.WriteString("- Add more seeds by editing `defaultOctoSeeds`.\n")

	return b.String()
}

func safeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

func buildMapSVG(host string, paths []string) string {
	uniq := map[string]bool{}
	clean := []string{}
	for _, p := range paths {
		p = "/" + strings.TrimLeft(strings.TrimSpace(p), "/")
		if p == "/" || p == "" {
			continue
		}
		if !uniq[p] {
			uniq[p] = true
			clean = append(clean, p)
		}
	}
	sort.Strings(clean)

	const (
		w      = 900
		lineH  = 18
		pad    = 18
		fontSz = 14
	)
	h := pad*2 + lineH*(len(clean)+2)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="` + itoa(w) + `" height="` + itoa(h) + `" viewBox="0 0 ` + itoa(w) + ` ` + itoa(h) + `">` + "\n")
	b.WriteString(`<rect x="0" y="0" width="100%" height="100%" fill="#0b0f19"/>` + "\n")
	b.WriteString(`<text x="` + itoa(pad) + `" y="` + itoa(pad+lineH) + `" font-family="monospace" font-size="` + itoa(fontSz+2) + `" fill="#e6edf3">` + escape(host) + `</text>` + "\n")
	y := pad + lineH*2
	for _, p := range clean {
		depth := strings.Count(strings.TrimLeft(p, "/"), "/")
		x := pad + depth*18
		label := "└ " + strings.TrimLeft(p, "/")
		b.WriteString(`<text x="` + itoa(x) + `" y="` + itoa(y) + `" font-family="monospace" font-size="` + itoa(fontSz) + `" fill="#a5b4fc">` + escape(label) + `</text>` + "\n")
		y += lineH
	}
	b.WriteString(`</svg>` + "\n")
	return b.String()
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [32]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func fmtAny(v any) string {
	switch t := v.(type) {
	case bool:
		if t {
			return "true"
		}
		return "false"
	case string:
		return t
	case []string:
		if len(t) == 0 {
			return "[]"
		}
		return "[" + strings.Join(t, ", ") + "]"
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
