#!/usr/bin/env bash
set -euo pipefail

echo "🔧 Restless OpenAPI compatibility fixer (one-shot)"

ROOT="$(pwd)"

fail() {
  echo "❌ $1"
  exit 1
}

echo "== 1. Fix report.go =="
cat > internal/modules/openapi/guard/report/report.go <<'EOF'
package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
)

func PrintHuman(res model.GuardResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Contract Drift Index: %.3f\n\n", res.CDI)

	if len(res.Findings) == 0 {
		b.WriteString("No contract violations detected.\n")
		return b.String()
	}

	for _, f := range res.Findings {
		fmt.Fprintf(&b,
			"%s %s %d %s [%s/%s]\n  %s\n\n",
			f.Method,
			f.Path,
			f.Status,
			f.JSONPath,
			f.Kind,
			f.Severity,
			f.Message,
		)
	}

	return b.String()
}

func ToJSON(res model.GuardResult) ([]byte, error) {
	return json.MarshalIndent(res, "", "  ")
}
EOF

echo "== 2. Fix loader.go =="
cat > internal/modules/openapi/guard/loader/loader.go <<'EOF'
package loader

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type LoadOptions struct {
	AllowRemoteRefs bool
}

func Load(ctx context.Context, ref string, opt LoadOptions) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = opt.AllowRemoteRefs
	loader.Context = ctx

	var doc *openapi3.T
	var err error

	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		resp, err := http.Get(ref)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, fmt.Errorf("openapi fetch failed: %s", resp.Status)
		}

		doc, err = loader.LoadFromReader(resp.Body)
	} else {
		if !filepath.IsAbs(ref) {
			ref, _ = filepath.Abs(ref)
		}
		doc, err = loader.LoadFromFile(ref)
	}

	if err != nil {
		return nil, err
	}

	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("openapi spec invalid: %w", err)
	}

	return doc, nil
}
EOF

echo "== 3. Replace Responses.Get with Map() =="

grep -rl 'Responses.Get' internal/modules/openapi/guard | while read -r f; do
  sed -i 's/op\.Responses\.Get(\([^)]*\))/op.Responses.Map()[\1]/g' "$f"
done

echo "== 4. Fix openapi3.Types return =="

grep -rl 'Schema.Value.Type' internal/modules/openapi/guard | while read -r f; do
  sed -i 's/return mt\.Schema\.Value\.Type/if mt.Schema.Value.Type != nil \&\& len(*mt.Schema.Value.Type) > 0 { return (*mt.Schema.Value.Type)[0] }\n\treturn ""/g' "$f" || true
done

echo "== 5. Fix jsonschema Causes loop =="

sed -i 's/if cv, ok := c.(.*); ok {/flattenValidationError(opID, method, path, status, contentType, c, out)\n\t\tcontinue\n\t}/g' \
  internal/modules/openapi/guard/runtime/validate_helpers.go || true

echo "== 6. go mod tidy =="
go mod tidy

echo "== 7. build test =="
if go build ./cmd/restless; then
  echo "✅ Build succeeded."
else
  fail "Build still failing. Inspect output above."
fi

echo "🎉 OpenAPI compatibility layer fixed."
