#!/usr/bin/env bash
set -euo pipefail

echo "🔧 Restless OpenAPI compatibility repair (stable edition)"

fail() {
  echo "❌ $1"
  exit 1
}

echo "== Fix loader.go (remove LoadFromReader, use LoadFromFile/URI) =="

cat > internal/modules/openapi/guard/loader/loader.go <<'EOF'
package loader

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type LoadOptions struct {
	AllowRemoteRefs bool
}

func Load(ctx context.Context, ref string, opt LoadOptions) (*openapi3.T, error) {
	ldr := openapi3.NewLoader()
	ldr.IsExternalRefsAllowed = opt.AllowRemoteRefs
	ldr.Context = ctx

	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		resp, err := http.Get(ref)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, fmt.Errorf("openapi fetch failed: %s", resp.Status)
		}

		tmp, err := os.CreateTemp("", "restless-openapi-*.json")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmp.Name())

		if _, err := tmp.ReadFrom(resp.Body); err != nil {
			return nil, err
		}
		tmp.Close()

		doc, err := ldr.LoadFromFile(tmp.Name())
		if err != nil {
			return nil, err
		}
		if err := doc.Validate(ctx); err != nil {
			return nil, err
		}
		return doc, nil
	}

	if !filepath.IsAbs(ref) {
		ref, _ = filepath.Abs(ref)
	}

	doc, err := ldr.LoadFromFile(ref)
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, err
	}

	return doc, nil
}
EOF

echo "== Fix validate_helpers.go safely =="

cat > internal/modules/openapi/guard/runtime/validate_helpers.go <<'EOF'
package runtime

import (
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
)

func mapSchemaError(opID, method, path string, status int, contentType string, err error) []model.Finding {
	var out []model.Finding

	if ve, ok := err.(*jsonschema.ValidationError); ok {
		flattenValidationError(opID, method, path, status, contentType, ve, &out)
		return out
	}

	out = append(out, model.Finding{
		OpID: opID,
		Method: strings.ToUpper(method),
		Path: path,
		Status: status,
		ContentType: contentType,
		Kind: model.KindSchemaViolation,
		Severity: model.SevHigh,
		JSONPath: "$",
		Message: err.Error(),
	})

	return out
}

func flattenValidationError(
	opID, method, path string,
	status int,
	contentType string,
	ve *jsonschema.ValidationError,
	out *[]model.Finding,
) {
	jp := "$"
	if ve.InstanceLocation != "" {
		jp = "$" + pointerToJSONPath(ve.InstanceLocation)
	}

	msg := ve.Message
	sev := model.SevHigh
	kind := model.KindSchemaViolation

	m := strings.ToLower(msg)

	switch {
	case strings.Contains(m, "required"):
		kind = model.KindMissingField
		sev = model.SevCritical
	case strings.Contains(m, "invalid type"), strings.Contains(m, "type"):
		kind = model.KindTypeMismatch
		sev = model.SevHigh
	case strings.Contains(m, "enum"):
		kind = model.KindEnumViolation
		sev = model.SevMedium
	}

	*out = append(*out, model.Finding{
		OpID: opID,
		Method: strings.ToUpper(method),
		Path: path,
		Status: status,
		ContentType: contentType,
		Kind: kind,
		Severity: sev,
		JSONPath: jp,
		Message: msg,
	})

	for _, c := range ve.Causes {
		flattenValidationError(opID, method, path, status, contentType, c, out)
	}
}

func pointerToJSONPath(ptr string) string {
	if ptr == "" || ptr == "/" {
		return ""
	}

	parts := strings.Split(ptr, "/")
	var b strings.Builder

	for _, p := range parts {
		if p == "" {
			continue
		}
		if isDigits(p) {
			b.WriteString("[")
			b.WriteString(p)
			b.WriteString("]")
		} else {
			b.WriteString(".")
			b.WriteString(p)
		}
	}
	return b.String()
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
EOF

echo "== go mod tidy =="
go mod tidy

echo "== rebuild =="
if go build ./cmd/restless; then
  echo "✅ Build succeeded."
else
  fail "Still failing."
fi

echo "🎉 OpenAPI compatibility fully repaired."
