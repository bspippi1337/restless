package engine

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var pathRegex = regexp.MustCompile(`https?://[^"]+`)

func CrawlAPI(target string) []string {

	target = normalizeTarget(target)

	client := &http.Client{
		Timeout: 6 * time.Second,
	}

	resp, err := client.Get(target)
	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	var body map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil
	}

	var endpoints []string
	seen := map[string]bool{}

	extractPaths(body, target, &endpoints, seen)

	return endpoints
}

func extractPaths(v interface{}, base string, out *[]string, seen map[string]bool) {

	switch val := v.(type) {

	case map[string]interface{}:

		for _, vv := range val {
			extractPaths(vv, base, out, seen)
		}

	case string:

		matches := pathRegex.FindAllString(val, -1)

		for _, m := range matches {

			u, err := url.Parse(m)
			if err != nil {
				continue
			}

			if u.Path == "" {
				continue
			}

			p := u.Path

			if !seen[p] {
				seen[p] = true
				*out = append(*out, p)
			}
		}
	}
}
