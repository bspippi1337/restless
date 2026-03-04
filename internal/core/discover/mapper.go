package discover

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Graph is the result of autonomous discovery.
type Graph struct {
	BaseURL   string
	Endpoints []string
	Visited   int
}

// Run crawls a hypermedia-style API surface starting from base URL.
// It extracts absolute http(s) URLs from JSON responses and follows same-host URLs.
// Intentionally bounded to avoid runaway crawling.
func Run(base string) (Graph, error) {
	client := http.Client{Timeout: 10 * time.Second}

	queue := []string{strings.TrimRight(base, "/")}
	seenURL := map[string]bool{}
	seenPath := map[string]bool{}

	var endpoints []string

	for len(queue) > 0 && len(seenURL) < 80 {
		u := queue[0]
		queue = queue[1:]

		if seenURL[u] {
			continue
		}
		seenURL[u] = true

		resp, err := client.Get(u)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var data interface{}
		_ = json.Unmarshal(body, &data)

		urls := extractURLs(data)

		for _, s := range urls {
			pu, err := url.Parse(s)
			if err != nil {
				continue
			}

			if pu.Path != "" && !seenPath[pu.Path] {
				seenPath[pu.Path] = true
				endpoints = append(endpoints, pu.Path)
			}

			if strings.Contains(pu.Host, hostOf(base)) {
				queue = append(queue, pu.String())
			}
		}
	}

	return Graph{
		BaseURL:   strings.TrimRight(base, "/"),
		Endpoints: endpoints,
		Visited:   len(seenURL),
	}, nil
}

func extractURLs(v interface{}) []string {
	var out []string

	switch x := v.(type) {
	case map[string]interface{}:
		for _, vv := range x {
			out = append(out, extractURLs(vv)...)
		}
	case []interface{}:
		for _, vv := range x {
			out = append(out, extractURLs(vv)...)
		}
	case string:
		s := strings.TrimSpace(x)
		if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
			out = append(out, s)
		}
	}

	return out
}

func hostOf(u string) string {
	p, _ := url.Parse(u)
	return p.Host
}
