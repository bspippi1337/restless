package engine

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Endpoint struct {
	Path       string
	Confidence string
	Type       string
}

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
	"/rate_limit",
	"/user",
	"/orgs",
}

func isUsefulStatus(code int) bool {
	return code == 200 || code == 201 || code == 202 || code == 204 || code == 401 || code == 403
}

func probeEndpoint(client *http.Client, target, path string) *Endpoint {
	url := target + path

	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if isUsefulStatus(resp.StatusCode) {
		return &Endpoint{
			Path:       path,
			Confidence: "medium",
		}
	}

	return nil
}

func CrawlEndpoint(target, path string) []string {
	target = normalizeTarget(target)

	client := &http.Client{
		Timeout: 6 * time.Second,
	}

	resp, err := client.Get(target + path)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "json") {
		return nil
	}

	return CrawlAPI(target + path)
}

func DiscoverEndpoints(target string) []Endpoint {
	target = normalizeTarget(target)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	seen := map[string]bool{}
	var endpoints []Endpoint
	var mu sync.Mutex

	pathChan := make(chan string)
	resultChan := make(chan *Endpoint)

	workers := 8
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range pathChan {
				res := probeEndpoint(client, target, p)
				if res != nil {
					resultChan <- res
				}
			}
		}()
	}

	go func() {
		for _, p := range commonPaths {
			pathChan <- p
		}
		close(pathChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for r := range resultChan {
		if !seen[r.Path] {
			seen[r.Path] = true
			endpoints = append(endpoints, *r)
		}
	}

	// root crawler
	for _, p := range CrawlAPI(target) {
		p = normalizeTemplate(p)
		if p == "" || p == "/" || isDocumentationPath(p) {
			continue
		}
		if !seen[p] {
			p = normalizeParameters(p)
			p = normalizeParameters(p)
			seen[p] = true
			endpoints = append(endpoints, Endpoint{
				Path:       p,
				Confidence: "high",
			})
		}
	}

	// recursive expansion, depth-limited
	queue := make([]string, 0, len(endpoints))
	for _, e := range endpoints {
		queue = append(queue, e.Path)
	}

	maxDepth := 2

	for depth := 0; depth < maxDepth; depth++ {
		var nextQueue []string
		var recWG sync.WaitGroup
		recChan := make(chan string, 256)

		for _, path := range queue {
			path := path
			recWG.Add(1)
			go func() {
				defer recWG.Done()
				for _, p := range CrawlEndpoint(target, path) {
					p = normalizeTemplate(p)
					if p == "" || p == "/" || isDocumentationPath(p) {
						continue
					}
					recChan <- p
				}
			}()
		}

		go func() {
			recWG.Wait()
			close(recChan)
		}()

		for p := range recChan {
			mu.Lock()
			if !seen[p] {
				p = normalizeParameters(p)
				p = normalizeParameters(p)
				seen[p] = true
				endpoints = append(endpoints, Endpoint{
					Path:       p,
					Confidence: "high",
				})
				nextQueue = append(nextQueue, p)
			}
			mu.Unlock()
		}

		if len(nextQueue) == 0 {
			break
		}
		queue = nextQueue
	}

	mergeTemplates(target, &endpoints, seen)

	inferResources(&endpoints, seen)

	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].Path < endpoints[j].Path
	})

	return endpoints
}
