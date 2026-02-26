#!/usr/bin/env bash
set -euo pipefail

echo "==> Hardening Round 1: Spec Validation"

cat >> internal/modules/openapi/run.go <<'EOT'

// Validate method/path exists
func ValidateEndpoint(spec Spec, method, path string) error {
	item, ok := spec.Paths[path]
	if !ok {
		return errors.New("path not found in spec")
	}
	if _, ok := item[strings.ToLower(method)]; !ok {
		return errors.New("method not allowed for this path")
	}
	return nil
}

// Check missing path params
func ValidatePathParams(path string, params map[string]string) error {
	for {
		start := strings.Index(path, "{")
		if start == -1 {
			break
		}
		end := strings.Index(path[start:], "}")
		if end == -1 {
			break
		}
		key := path[start+1 : start+end]
		if _, ok := params[key]; !ok {
			return fmt.Errorf("missing path param: %s", key)
		}
		path = path[start+end+1:]
	}
	return nil
}
EOT

echo "==> gofmt/build"
gofmt -w internal/modules/openapi/run.go
go build -o restless-v2 ./cmd/restless-v2

echo "✅ Round 1 installed"
