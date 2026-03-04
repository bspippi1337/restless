package discover

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Graph is the result of autonomous discovery.
type Graph struct {
	BaseURL   string
	Endpoints []string // paths only, deduped
	Visited   int
	Notes     []string
}

// Options tune the crawler.
type Options struct {
	Timeout       time.Duration
	MaxRequests   int
	MaxQueue      int
	MaxPerPath    int
	MaxDepth      int
	UserAgent     string
	FollowLinkHdr bool
	ExpandTmpl    bool
}

func DefaultOptions() Options {
	return Options{
		Timeout:       12 * time.Second,
		MaxRequests:   60,
		MaxQueue:      300,
		MaxPerPath:    3,
		MaxDepth:      3,
		UserAgent:     "restless-discover/1 (+https://github.com/bspippi1337/restless)",
		FollowLinkHdr: true,
		ExpandTmpl:    true,
	}
}

type queued struct {
	u     string
	depth int
}

func Run(base string) (Graph, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return RunWith(ctx, base, DefaultOptions())
}

func RunWith(ctx context.Context, base string, opt Options) (Graph, error) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return Graph{}, errors.New("empty base url")
	}
	baseURL, err := url.Parse(base)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return Graph{}, fmt.Errorf("invalid url: %q", base)
	}

	client := &http.Client{Timeout: opt.Timeout}

	seenURL := map[string]bool{}
	seenPath := map[string]bool{}
	pathHits := map[string]int{}
	var endpoints []string
	var notes []string

	// Seed queue with base itself + likely OpenAPI locations (we don't parse spec fully yet,
	// but their JSON often includes more URLs or paths).
	queue := []queued{
		{u: base, depth: 0},
		{u: base + "/openapi.json", depth: 0},
		{u: base + "/swagger.json", depth: 0},
		{u: base + "/v3/api-docs", depth: 0},
	}

	reqCount := 0

	pushURL := func(u string, depth int) {
		if len(queue) >= opt.MaxQueue {
			return
		}
		if depth > opt.MaxDepth {
			return
		}
		if seenURL[u] {
			return
		}
		queue = append(queue, queued{u: u, depth: depth})
	}

	addPath := func(p string) {
		if p == "" {
			return
		}
		if !strings.HasPrefix(p, "/") {
			return
		}
		// prevent runaway due to templates expanding into huge strings
		if len(p) > 200 {
			return
		}
		if seenPath[p] {
			return
		}
		seenPath[p] = true
		endpoints = append(endpoints, p)
	}

	sameHost := func(u *url.URL) bool {
		return strings.EqualFold(u.Host, baseURL.Host)
	}

	for len(queue) > 0 && reqCount < opt.MaxRequests {
		item := queue[0]
		queue = queue[1:]

		if seenURL[item.u] {
			continue
		}
		seenURL[item.u] = true

		reqCount++

		req, _ := http.NewRequestWithContext(ctx, "GET", item.u, nil)
		req.Header.Set("User-Agent", opt.UserAgent)
		req.Header.Set("Accept", "application/json, application/vnd.github+json;q=0.9, */*;q=0.1")

		resp, err := client.Do(req)
		if err != nil || resp == nil {
			continue
		}

		// Always close body
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB max
		resp.Body.Close()

		// Record path
		if u, err := url.Parse(item.u); err == nil && u.Path != "" {
			addPath(u.Path)
			pathHits[u.Path]++
			if pathHits[u.Path] > opt.MaxPerPath {
				// avoid hammering the same path via different query strings
				continue
			}
		}

		// Link header pagination
		if opt.FollowLinkHdr {
			for _, nu := range parseLinkHeader(resp.Header.Get("Link")) {
				if nu == "" {
					continue
				}
				pu, err := url.Parse(nu)
				if err != nil || pu.Scheme == "" || pu.Host == "" {
					continue
				}
				if sameHost(pu) {
					pushURL(pu.String(), item.depth+1)
				}
			}
		}

		// Only attempt JSON extraction if it looks like JSON
		ct := resp.Header.Get("Content-Type")
		looksJSON := strings.Contains(ct, "json") || bytes.HasPrefix(bytes.TrimSpace(body), []byte("{")) || bytes.HasPrefix(bytes.TrimSpace(body), []byte("["))
		if !looksJSON || len(body) == 0 {
			continue
		}

		var data any
		if err := json.Unmarshal(body, &data); err != nil {
			continue
		}

		// Extract absolute URLs from JSON
		absURLs := extractAbsoluteURLs(data)
		for _, s := range absURLs {
			pu, err := url.Parse(s)
			if err != nil || pu.Scheme == "" || pu.Host == "" {
				continue
			}
			if !sameHost(pu) {
				continue
			}
			addPath(pu.Path)

			// Expand templates and queue them (bounded)
			if opt.ExpandTmpl && strings.Contains(pu.Path, "{") {
				for _, ex := range expandTemplatePath(pu.Path) {
					addPath(ex)
					nu := *pu
					nu.Path = ex
					nu.RawQuery = "" // avoid infinite query variants
					pushURL(nu.String(), item.depth+1)
				}
			}

			// Queue discovered URLs (bounded)
			pu.RawQuery = "" // avoid pagination query explosion; Link header handles most pagination
			pushURL(pu.String(), item.depth+1)
		}
	}

	sort.Strings(endpoints)

	// Notes
	if reqCount >= opt.MaxRequests {
		notes = append(notes, fmt.Sprintf("hit MaxRequests=%d (increase if needed)", opt.MaxRequests))
	}
	if len(queue) >= opt.MaxQueue {
		notes = append(notes, fmt.Sprintf("hit MaxQueue=%d (increase if needed)", opt.MaxQueue))
	}
	if len(endpoints) == 0 {
		notes = append(notes, "no endpoints discovered (target may not expose hypermedia/root JSON)")
	}

	return Graph{
		BaseURL:   base,
		Endpoints: endpoints,
		Visited:   len(seenURL),
		Notes:     notes,
	}, nil
}

// extractAbsoluteURLs walks arbitrary JSON and returns strings that look like absolute http(s) URLs.
func extractAbsoluteURLs(v any) []string {
	var out []string
	switch t := v.(type) {
	case map[string]any:
		for _, vv := range t {
			out = append(out, extractAbsoluteURLs(vv)...)
		}
	case []any:
		for _, vv := range t {
			out = append(out, extractAbsoluteURLs(vv)...)
		}
	case string:
		s := strings.TrimSpace(t)
		if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
			out = append(out, s)
		}
	}
	return out
}

// parseLinkHeader extracts rel="next" (and any) URLs from an RFC5988-ish Link header.
// Example: <https://api.github.com/...>; rel="next", <...>; rel="last"
func parseLinkHeader(h string) []string {
	if h == "" {
		return nil
	}
	parts := strings.Split(h, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if !strings.HasPrefix(p, "<") {
			continue
		}
		end := strings.Index(p, ">")
		if end <= 1 {
			continue
		}
		u := p[1:end]
		out = append(out, u)
	}
	return out
}

// expandTemplatePath generates a small set of safe expansions for common template tokens.
// Keeps it intentionally conservative.
func expandTemplatePath(path string) []string {
	// Replace {owner}/{repo}/{user}/{org}/{id} etc with plausible tokens.
	repl := func(s string) string {
		s = strings.ReplaceAll(s, "{owner}", "octocat")
		s = strings.ReplaceAll(s, "{repo}", "hello-world")
		s = strings.ReplaceAll(s, "{user}", "octocat")
		s = strings.ReplaceAll(s, "{org}", "github")
		s = strings.ReplaceAll(s, "{id}", "1")
		s = strings.ReplaceAll(s, "{gist_id}", "1")
		s = strings.ReplaceAll(s, "{name}", "test")
		// Any remaining {x} -> "1"
		for {
			i := strings.Index(s, "{")
			j := strings.Index(s, "}")
			if i == -1 || j == -1 || j < i {
				break
			}
			s = s[:i] + "1" + s[j+1:]
		}
		return s
	}

	a := repl(path)
	if a == path {
		return nil
	}

	// Also include a “me” style variant for some APIs
	b := strings.ReplaceAll(a, "/1", "/me")

	uniq := map[string]bool{a: true, b: true}
	var out []string
	for k := range uniq {
		// avoid nonsense root-only expansions
		if k != "" && strings.HasPrefix(k, "/") {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
