package auto

import (
	"net/http"
	"time"
)

var commonPaths = []string{
	"/openapi.json",
	"/swagger.json",
	"/v3/api-docs",
	"/openapi.yaml",
	"/swagger.yaml",
}

// TryDiscover tries common OpenAPI/Swagger endpoints under baseURL.
// baseURL should be like: https://api.example.com (no trailing slash preferred).
func TryDiscover(baseURL string) (string, bool) {
	client := http.Client{Timeout: 4 * time.Second}

	for _, p := range commonPaths {
		url := baseURL + p
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 200 {
			return url, true
		}
	}
	return "", false
}
