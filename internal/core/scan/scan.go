package scan

import (
	"context"
	"net/http"
	"time"

	"github.com/bspippi1337/restless/internal/core/state"
)

func Run(ctx context.Context, base string) (state.ScanResult, error) {

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	hints := []string{
		"/",
		"/health",
		"/api",
		"/v1",
	}

	var routes []state.Route

	for _, p := range hints {

		req, _ := http.NewRequestWithContext(ctx, "GET", base+p, nil)

		resp, err := client.Do(req)
		if err == nil && resp != nil {

			resp.Body.Close()

			if resp.StatusCode != 404 {
				routes = append(routes, state.Route{
					Method: "GET",
					Path:   p,
				})
			}
		}
	}

	return state.ScanResult{
		BaseURL:   base,
		Endpoints: routes,
	}, nil
}
