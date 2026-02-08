package scrape

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

type PathHit struct {
	Method string
	Path   string
}

func LightDocsScrape(ctx context.Context, root string, budgetPages int) ([]PathHit, []string) {
	cands := []string{
		root,
		root + "/docs",
		root + "/documentation",
		root + "/api",
		root + "/developers",
	}
	if budgetPages < 1 { budgetPages = 1 }
	if budgetPages > len(cands) { budgetPages = len(cands) }
	cands = cands[:budgetPages]

	client := &http.Client{}
	paths := map[string]struct{}{}
	visited := []string{}

	rePath := regexp.MustCompile(`(?i)(/v\d+/(?:[a-z0-9_\-]+/?)+)|(/api/(?:[a-z0-9_\-]+/?)+)|(/(?:health|status|version)(?:\b|/))`)

	for _, u := range cands {
		req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
		req.Header.Set("User-Agent", "restless-discovery/0.2")
		res, err := client.Do(req)
		if err != nil { continue }
		visited = append(visited, u)
		b, _ := io.ReadAll(io.LimitReader(res.Body, 2<<20))
		res.Body.Close()

		m := rePath.FindAllString(string(b), -1)
		for _, p := range m {
			pp := strings.TrimSpace(p)
			if pp == "" || !strings.HasPrefix(pp, "/") { continue }
			paths[pp] = struct{}{}
		}
	}

	out := []PathHit{}
	for p := range paths {
		out = append(out, PathHit{Method: "GET", Path: p})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, visited
}
