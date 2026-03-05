#!/usr/bin/env bash
set -euo pipefail

echo "==> Fixing session flow module"

cat > internal/modules/session/flow.go <<'EOT'
package session

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

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

func RunFlow(ctx context.Context, a *app.App, steps []FlowStep, sess *Module) error {
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

		for varName, path := range s.Extract {
			val, err := extractDot(path, resp.Body)
			if err != nil {
				return err
			}
			sess.Set(varName, val)
		}
	}
	return nil
}

func extractDot(path string, body []byte) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}

	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return "", err
	}

	cur := v
	parts := strings.Split(path, ".")
	for _, part := range parts {
		obj, ok := cur.(map[string]any)
		if !ok {
			return "", errors.New("path not found")
		}
		next, ok := obj[part]
		if !ok {
			return "", errors.New("path not found")
		}
		cur = next
	}

	switch t := cur.(type) {
	case string:
		return t, nil
	default:
		b, _ := json.Marshal(t)
		return string(b), nil
	}
}

func mapToHeader(m map[string]string) map[string][]string {
	h := map[string][]string{}
	for k, v := range m {
		h[k] = []string{v}
	}
	return h
}
EOT

echo "==> Formatting + testing"
gofmt -w internal/modules/session
go test ./...

echo "Session flow fixed."
