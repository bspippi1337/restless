#!/usr/bin/env bash
set -euo pipefail

echo "==> Adding Flow Runner"

mkdir -p internal/modules/session

cat > internal/modules/session/flow.go <<'EOT'
package session

import (
	"context"
	"encoding/json"
	"os"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/core/types"
)

type FlowStep struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Extract map[string]string `json:"extract"` // var -> json dot path
}

func RunFlow(ctx context.Context, a *app.App, steps []FlowStep) error {
	for _, s := range steps {
		req := types.Request{
			Method:  s.Method,
			URL:     s.URL,
			Headers: mapToHeader(s.Headers),
			Body:    []byte(s.Body),
		}
		resp, err := a.RunOnce(ctx, req)
		if err != nil {
			return err
		}

		if len(s.Extract) > 0 {
			for k, path := range s.Extract {
				val, err := extractDot(path, resp.Body)
				if err == nil {
					// naive: write to env file for now
					_ = os.WriteFile(".restless/last_extract.json",
						mustJSON(map[string]string{k: val}),
						0644)
				}
			}
		}
	}
	return nil
}

func extractDot(path string, body []byte) (string, error) {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return "", err
	}
	cur := v
	for _, part := range split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return "", os.ErrInvalid
		}
		cur = m[part]
	}
	b, _ := json.Marshal(cur)
	return string(b), nil
}

func split(s, sep string) []string {
	var out []string
	for _, p := range []rune(s) {
		_ = p
	}
	return []string{} // placeholder simple; expand later
}

func mapToHeader(m map[string]string) map[string][]string {
	h := map[string][]string{}
	for k, v := range m {
		h[k] = []string{v}
	}
	return h
}

func mustJSON(v any) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return b
}
EOT

echo "==> Enhancing Bench Output"

cat > internal/modules/bench/table.go <<'EOT'
package bench

import "fmt"

func PrintTable(r Result) {
	fmt.Println("==== BENCH RESULT ====")
	fmt.Printf("Total     : %d\n", r.TotalRequests)
	fmt.Printf("Errors    : %d\n", r.Errors)
	fmt.Printf("Duration  : %d ms\n", r.DurationMs)
	fmt.Printf("P50       : %d ms\n", r.P50Ms)
	fmt.Printf("P95       : %d ms\n", r.P95Ms)
	fmt.Printf("P99       : %d ms\n", r.P99Ms)
}
EOT

echo "==> Adding OpenAPI CLI commands"

cat > internal/modules/openapi/cli.go <<'EOT'
package openapi

import (
	"fmt"
)

func ListCached() error {
	fmt.Println("OpenAPI cache listing not yet implemented (v1 skeleton).")
	return nil
}
EOT

gofmt -w .
go test ./...

echo "Phase 2.5 applied."
