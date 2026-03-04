package discover

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Graph struct {
	BaseURL   string
	Endpoints []string
	Visited   int
}

func Run(base string) (Graph, error) {

	client := http.Client{Timeout: 10 * time.Second}

	queue := []string{base}
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
		json.Unmarshal(body, &data)

		urls := extract(data)

		for _, s := range urls {

			p, err := url.Parse(s)
			if err != nil {
				continue
			}

			path := p.Path

			if path != "" && !seenPath[path] {
				seenPath[path] = true
				endpoints = append(endpoints, path)
			}

			if strings.Contains(p.Host, host(base)) {
				queue = append(queue, s)
			}
		}
	}

	return Graph{
		BaseURL:   base,
		Endpoints: endpoints,
		Visited:   len(seenURL),
	}, nil
}

func extract(v interface{}) []string {

	var out []string

	switch x := v.(type) {

	case map[string]interface{}:
		for _, v := range x {
			out = append(out, extract(v)...)
		}

	case []interface{}:
		for _, v := range x {
			out = append(out, extract(v)...)
		}

	case string:
		if strings.HasPrefix(x, "http") {
			out = append(out, x)
		}
	}

	return out
}

func host(u string) string {
	p, _ := url.Parse(u)
	return p.Host
}