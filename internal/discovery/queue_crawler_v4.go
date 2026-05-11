package discovery

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"

	"github.com/bspippi1337/restless/internal/httpx"
	"github.com/bspippi1337/restless/internal/store"
	"github.com/bspippi1337/restless/internal/telemetry"
	"github.com/bspippi1337/restless/internal/util"
)

type job struct {
	path string
}

func CrawlQueueV4(base string, workers int) []store.Endpoint {

	client := httpx.New()

	queue := make(chan job, 256)

	var endpoints []store.Endpoint

	seen := map[string]bool{}
	var mu sync.Mutex

	var wg sync.WaitGroup

	enqueue := func(p string) {

		if p == "" {
			return
		}

		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}

		mu.Lock()
		if seen[p] {
			mu.Unlock()
			return
		}

		seen[p] = true
		mu.Unlock()

		wg.Add(1)

		queue <- job{p}
	}

	enqueue("/")

	for i := 0; i < workers; i++ {

		go func() {

			for j := range queue {
				func() {
					defer wg.Done()
					url := util.JoinURL(base, j.path)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
					telemetry.IncRequest()
					res, err := client.HTTP.Do(req)
					cancel()
					if err != nil {
						return
					}
					body, _ := io.ReadAll(res.Body)
					res.Body.Close()
					if res.StatusCode < 400 {
						telemetry.IncEndpoint()
						mu.Lock()
						endpoints = append(endpoints, store.Endpoint{Path: j.path})
						mu.Unlock()
					}
					var obj any
					if json.Unmarshal(body, &obj) != nil {
						return
					}
					var walk func(any)
					walk = func(v any) {
						switch t := v.(type) {
						case map[string]any:
							for _, x := range t {
								walk(x)
							}
						case []any:
							for _, x := range t {
								walk(x)
							}
						case string:
							s := strings.TrimSpace(t)
							if strings.HasPrefix(s, "http") {
								u, err := neturl.Parse(s)
								if err == nil && strings.Contains(base, u.Host) {
									enqueue(u.Path)
								}
							}
							if strings.HasPrefix(s, "/") {
								enqueue(s)
							}
						}
					}
					walk(obj)
				}()

			}

		}()

	}

	go func() {
		wg.Wait()
		close(queue)
	}()

	wg.Wait()

	return endpoints
}
