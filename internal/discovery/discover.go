package discovery

import (
	"net/http"
	"strings"
	"time"
)

type Result struct {
	BaseURL   string     `json:"base_url"`
	Endpoints []Endpoint `json:"endpoints"`
}

func Discover(url string) (*Result, error) {
	client := &http.Client{
		Timeout: 12 * time.Second,
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	endpoints := []Endpoint{
		{Path: "/"},
	}

	guesses := []string{
		"/health",
		"/healthz",
		"/status",
		"/api",
		"/api/v1",
		"/swagger",
		"/swagger.json",
		"/openapi.json",
		"/docs",
	}

	for _, p := range guesses {
		req, err := http.NewRequest(http.MethodHead, strings.TrimRight(url, "/")+p, nil)
		if err != nil {
			continue
		}
		r, err := client.Do(req)
		if err != nil {
			continue
		}
		r.Body.Close()
		if r.StatusCode > 0 && r.StatusCode < 500 {
			endpoints = append(endpoints, Endpoint{Path: p})
		}
	}

	return &Result{
		BaseURL:   url,
		Endpoints: dedupeEndpoints(endpoints),
	}, nil
}

func dedupeEndpoints(in []Endpoint) []Endpoint {
	seen := make(map[string]struct{}, len(in))
	out := make([]Endpoint, 0, len(in))
	for _, e := range in {
		if e.Path == "" {
			continue
		}
		if _, ok := seen[e.Path]; ok {
			continue
		}
		seen[e.Path] = struct{}{}
		out = append(out, e)
	}
	return out
}
