package discover

import (
	"net/http"
	"strings"
	"time"
)

type Graph struct {
	BaseURL   string
	Endpoints []string
}

func Run(base string) (Graph, error) {

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	seeds := []string{
		"/",
		"/api",
		"/v1",
		"/v2",
		"/openapi.json",
		"/swagger.json",
	}

	seen := map[string]bool{}
	var endpoints []string

	for _, s := range seeds {

		u := base + s

		resp, err := client.Get(u)
		if err != nil {
			continue
		}

		resp.Body.Close()

		if resp.StatusCode != 404 {

			if !seen[s] {
				seen[s] = true
				endpoints = append(endpoints, s)
			}

			links := guessLinks(s)

			for _, l := range links {

				if !seen[l] {
					seen[l] = true
					endpoints = append(endpoints, l)
				}
			}
		}
	}

	return Graph{
		BaseURL:   base,
		Endpoints: endpoints,
	}, nil
}

func guessLinks(path string) []string {

	var out []string

	parts := strings.Split(path, "/")

	for i := range parts {

		if parts[i] == "" {
			continue
		}

		out = append(out, "/"+parts[i])
	}

	return out
}
