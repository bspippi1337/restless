package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bspippi1337/restless/internal/httpx"
	"github.com/bspippi1337/restless/internal/status"
	"github.com/bspippi1337/restless/internal/store"
	"github.com/bspippi1337/restless/internal/util"
)

type Engine struct {
	base   string
	client *httpx.Client
	seen   map[string]bool
	mu     sync.Mutex
}

func NewEngine(base string) (*Engine, error) {

	_, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	return &Engine{
		base:   base,
		client: httpx.New(),
		seen:   map[string]bool{},
	}, nil
}

func (e *Engine) Discover(ctx context.Context) []store.Endpoint {

	queue := make(chan string, 100)
	results := []store.Endpoint{}

	var wg sync.WaitGroup

	workers := 6

	queue <- "/"

	for i := 0; i < workers; i++ {

		wg.Add(1)

		go func() {
			defer wg.Done()

			for {

				select {

				case path := <-queue:

					e.mu.Lock()
					if e.seen[path] {
						e.mu.Unlock()
						continue
					}
					e.seen[path] = true
					e.mu.Unlock()

					status.IncRequest()

					full := util.JoinURL(e.base, path)

					req, _ := http.NewRequestWithContext(ctx, "GET", full, nil)

					res, err := e.client.HTTP.Do(req)
					if err != nil {
						status.IncError()
						continue
					}

					status.IncRequest()

					if res.StatusCode >= 400 {
						res.Body.Close()
						continue
					}

					results = append(results, store.Endpoint{
						Path: path,
					})

					status.IncEndpoint()

					var body map[string]interface{}

					json.NewDecoder(res.Body).Decode(&body)
					res.Body.Close()

					for _, v := range body {

						s, ok := v.(string)
						if !ok {
							continue
						}

						if !strings.Contains(s, "/") {
							continue
						}

						u, err := url.Parse(s)
						if err != nil {
							continue
						}

						p := u.Path

						if !strings.HasPrefix(p, "/") {
							continue
						}

						select {
						case queue <- p:
						default:
						}

					}

				case <-ctx.Done():
					return

				}

			}
		}()

	}

	time.Sleep(3 * time.Second)

	close(queue)

	wg.Wait()

	return results
}

func (e *Engine) PrintMap() {

	fmt.Println()

	for p := range e.seen {
		fmt.Println(" ", p)
	}

}
