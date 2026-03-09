package engine

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

var commonPaths = []string{
	"/api",
	"/api/v1",
	"/api/v2",
	"/users",
	"/repos",
	"/issues",
	"/search",
	"/graphql",
	"/health",
	"/status",
}

func normalize(target string) string {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	return "https://" + target
}

func isUsefulStatus(code int) bool {
	return code == 200 || code == 201 || code == 202 || code == 204 || code == 401 || code == 403
}

func DiscoverEndpoints(target string) []string {

	target = normalize(target)

	client := &http.Client{
		Timeout: 4 * time.Second,
	}

	seen := map[string]bool{}
	var endpoints []string

	for _, p := range commonPaths {

		url := target + p

		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		if isUsefulStatus(resp.StatusCode) && !seen[p] {
			seen[p] = true
			endpoints = append(endpoints, p)
		}

		resp.Body.Close()
	}

	sort.Strings(endpoints)
	return endpoints
}
