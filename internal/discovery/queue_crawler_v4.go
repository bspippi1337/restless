package discovery

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"sync/atomic"
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

	var inflight int64

	enqueue := func(p string) {

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

		atomic.AddInt64(&inflight, 1)

		queue <- job{p}
	}

	enqueue("/")

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {

		wg.Add(1)

		go func() {

			defer wg.Done()

			for j := range queue {

				url := util.JoinURL(base, j.path)

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

				telemetry.IncRequest()

				res, err := client.HTTP.Do(req)

				cancel()

				if err == nil {

					body, _ := io.ReadAll(res.Body)
					res.Body.Close()

					if res.StatusCode < 400 {

						telemetry.IncEndpoint()

						mu.Lock()
						endpoints = append(endpoints, store.Endpoint{Path: j.path})
						mu.Unlock()

					}

					var obj any

					if json.Unmarshal(body, &obj) == nil {

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

								if strings.HasPrefix(s, "/") {

									enqueue(s)

								}

								if strings.HasPrefix(s, base) {

									u, err := neturl.Parse(s)

									if err == nil {
										enqueue(u.Path)
									}

								}

							}

						}

						walk(obj)

					}

				}

				atomic.AddInt64(&inflight, -1)

			}

		}()

	}

	go func() {

		for {

			time.Sleep(100 * time.Millisecond)

			if atomic.LoadInt64(&inflight) == 0 {
				close(queue)
				return
			}

		}

	}()

	wg.Wait()

	return endpoints
}
