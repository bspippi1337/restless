package docparse

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type OpenAPI struct {
	OpenAPI string                    `json:"openapi,omitempty" yaml:"openapi,omitempty"`
	Swagger string                    `json:"swagger,omitempty" yaml:"swagger,omitempty"`
	Paths   map[string]map[string]any `json:"paths,omitempty" yaml:"paths,omitempty"`
}

func TryOpenAPI(ctx context.Context, root string) (*OpenAPI, []string, error) {
	cands := []string{
		root + "/openapi.json",
		root + "/swagger.json",
		root + "/api-docs",
		root + "/v1/openapi.json",
		root + "/.well-known/openapi.json",
		root + "/openapi.yaml",
		root + "/openapi.yml",
		root + "/.well-known/openapi.yaml",
		root + "/.well-known/openapi.yml",
	}
	client := &http.Client{}
	var lastErr error
	for _, u := range cands {
		req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
		req.Header.Set("Accept", "application/json, application/yaml, text/yaml, */*")
		res, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		b, err := io.ReadAll(io.LimitReader(res.Body, 12<<20))
		res.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			lastErr = errors.New(res.Status)
			continue
		}
		txt := strings.TrimSpace(string(b))
		if txt == "" {
			lastErr = errors.New("empty")
			continue
		}
		var o OpenAPI
		if err := yaml.Unmarshal([]byte(txt), &o); err != nil {
			lastErr = err
			continue
		}
		if len(o.Paths) == 0 {
			lastErr = errors.New("openapi has no paths")
			continue
		}
		return &o, []string{u}, nil
	}
	if lastErr == nil {
		lastErr = errors.New("not found")
	}
	return nil, nil, lastErr
}
