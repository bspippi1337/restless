#!/usr/bin/env bash
set -euo pipefail

# Add Restless "validate" MVP command (OpenAPI -> live checks)
# Run from repo root (where go.mod exists)

die(){ echo "ERROR: $*" >&2; exit 1; }
have(){ command -v "$1" >/dev/null 2>&1; }

[ -f go.mod ] || die "Run from repo root (missing go.mod)."
have go || die "go not found"
have git || die "git not found"

mkdir -p internal/validate cmd/restless

# --- Go dependency (kin-openapi) ---
echo "==> Ensuring dependency github.com/getkin/kin-openapi/openapi3 ..."
go get github.com/getkin/kin-openapi/openapi3@latest >/dev/null
go mod tidy >/dev/null

# --- internal validator ---
cat > internal/validate/validate.go <<'GO'
package validate

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

type Options struct {
	SpecPath   string
	BaseURL    string
	Timeout    time.Duration
	StrictLive bool   // if true: 404 is failure; if false: allow 401/403 but still flag 404
	AuthHeader string // e.g. "Authorization: Bearer XXX" (optional)
	JSON       bool
}

type Finding struct {
	Method        string `json:"method"`
	Path          string `json:"path"`
	URL           string `json:"url"`
	ExpectedCodes string `json:"expectedCodes"`
	ActualCode    int    `json:"actualCode"`
	Problem       string `json:"problem"`
}

type Report struct {
	OK       bool      `json:"ok"`
	BaseURL  string    `json:"baseUrl"`
	SpecPath string    `json:"specPath"`
	Checked  int       `json:"checked"`
	Failed   int       `json:"failed"`
	Findings []Finding `json:"findings"`
}

func Run(ctx context.Context, opt Options) (Report, error) {
	if opt.SpecPath == "" || opt.BaseURL == "" {
		return Report{}, errors.New("missing --spec or --base")
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 15 * time.Second
	}

	spec, err := loadSpec(ctx, opt.SpecPath)
	if err != nil {
		return Report{}, err
	}

	base, err := url.Parse(opt.BaseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return Report{}, fmt.Errorf("invalid --base: %q", opt.BaseURL)
	}

	client := &http.Client{
		Timeout: opt.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}

	var findings []Finding
	checked := 0

	// Iterate paths + operations
	for path, item := range spec.Paths.Map() {
		ops := operations(item)
		for method, op := range ops {
			checked++

			u := *base
			u.Path = joinURLPath(base.Path, materializePath(path))

			exp := expectedCodes(op)
			code, problem := hit(ctx, client, method, u.String(), opt.AuthHeader)

			// Core rule: 404 is drift (endpoint missing)
			// In non-strict mode we don't fail on 401/403 (auth required), but we still report mismatched codes
			fail := false
			if code == 404 {
				fail = true
			} else if opt.StrictLive {
				// strict: must match one of the response codes if spec has explicit codes
				if exp != "" && !codeMatchesExpected(code, exp) {
					fail = true
				}
			} else {
				// non-strict: allow auth blockers, but still flag obvious drift
				if exp != "" && code != 401 && code != 403 && !codeMatchesExpected(code, exp) {
					fail = true
				}
			}

			if fail {
				findings = append(findings, Finding{
					Method:        method,
					Path:          path,
					URL:           u.String(),
					ExpectedCodes: exp,
					ActualCode:    code,
					Problem:       problemOrDefault(problem, "drift detected"),
				})
			}
		}
	}

	rep := Report{
		OK:       len(findings) == 0,
		BaseURL:  opt.BaseURL,
		SpecPath: opt.SpecPath,
		Checked:  checked,
		Failed:   len(findings),
		Findings: findings,
	}

	return rep, nil
}

func loadSpec(ctx context.Context, path string) (*openapi3.T, error) {
	ldr := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, err := ldr.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("openapi load: %w", err)
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("openapi validate: %w", err)
	}
	return doc, nil
}

func operations(item *openapi3.PathItem) map[string]*openapi3.Operation {
	out := map[string]*openapi3.Operation{}
	if item.Get != nil {
		out["GET"] = item.Get
	}
	if item.Post != nil {
		out["POST"] = item.Post
	}
	if item.Put != nil {
		out["PUT"] = item.Put
	}
	if item.Patch != nil {
		out["PATCH"] = item.Patch
	}
	if item.Delete != nil {
		out["DELETE"] = item.Delete
	}
	if item.Head != nil {
		out["HEAD"] = item.Head
	}
	if item.Options != nil {
		out["OPTIONS"] = item.Options
	}
	return out
}

var rePathParam = regexp.MustCompile(`\{[^/]+\}`)

func materializePath(p string) string {
	// Replace {id} -> 1 (safe placeholder)
	p = rePathParam.ReplaceAllString(p, "1")
	// Replace :id -> 1 (some APIs)
	p = regexp.MustCompile(`:[A-Za-z_][A-Za-z0-9_]*`).ReplaceAllString(p, "1")
	return p
}

func joinURLPath(a, b string) string {
	if a == "" {
		return b
	}
	if strings.HasSuffix(a, "/") && strings.HasPrefix(b, "/") {
		return a + strings.TrimPrefix(b, "/")
	}
	if !strings.HasSuffix(a, "/") && !strings.HasPrefix(b, "/") {
		return a + "/" + b
	}
	return a + b
}

func expectedCodes(op *openapi3.Operation) string {
	if op == nil || op.Responses == nil {
		return ""
	}
	// Collect explicit numeric codes + patterns like "2XX"
	var codes []string
	for code := range op.Responses.Map() {
		codes = append(codes, code)
	}
	// Render stable-ish
	if len(codes) == 0 {
		return ""
	}
	// Keep it readable
	return strings.Join(codes, ",")
}

func codeMatchesExpected(code int, expected string) bool {
	// expected is comma-separated OpenAPI response keys: "200,201,default,2XX"
	parts := strings.Split(expected, ",")
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToUpper(p))
		if p == "" {
			continue
		}
		if p == "DEFAULT" {
			// default means any code is acceptable in spec terms
			return true
		}
		if strings.HasSuffix(p, "XX") && len(p) == 3 {
			// 2XX, 4XX etc
			d := int(p[0] - '0')
			if code/100 == d {
				return true
			}
		}
		// numeric
		if n := atoiSafe(p); n > 0 && n == code {
			return true
		}
	}
	return false
}

func atoiSafe(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return -1
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func hit(ctx context.Context, client *http.Client, method, target, authHeader string) (int, string) {
	req, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		return 0, "request build failed"
	}

	// Optional single header "Key: Value"
	if authHeader != "" {
		if k, v, ok := strings.Cut(authHeader, ":"); ok {
			req.Header.Set(strings.TrimSpace(k), strings.TrimSpace(v))
		}
	}

	// Give servers something sane
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "request failed: " + err.Error()
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, ""
}

func problemOrDefault(p, d string) string {
	if strings.TrimSpace(p) == "" {
		return d
	}
	return p
}

func PrintHuman(rep Report, w io.Writer) {
	if rep.OK {
		fmt.Fprintf(w, "✔ validate OK\n")
		fmt.Fprintf(w, "  base: %s\n  spec: %s\n  checked: %d\n", rep.BaseURL, rep.SpecPath, rep.Checked)
		return
	}
	fmt.Fprintf(w, "✖ validate FAILED\n")
	fmt.Fprintf(w, "  base: %s\n  spec: %s\n  checked: %d  failed: %d\n\n", rep.BaseURL, rep.SpecPath, rep.Checked, rep.Failed)
	for _, f := range rep.Findings {
		fmt.Fprintf(w, "- %s %s\n  url: %s\n  expected: %s\n  got: %d\n  problem: %s\n\n",
			f.Method, f.Path, f.URL, f.ExpectedCodes, f.ActualCode, f.Problem)
	}
}

func PrintJSON(rep Report, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

// Convenience: allow token via env if user doesn't pass header explicitly.
func AuthHeaderFromEnv() string {
	// RESTLESS_AUTH="Authorization: Bearer XXX"
	if v := strings.TrimSpace(os.Getenv("RESTLESS_AUTH")); v != "" {
		return v
	}
	// RESTLESS_TOKEN="XXX" -> Authorization: Bearer XXX
	if t := strings.TrimSpace(os.Getenv("RESTLESS_TOKEN")); t != "" {
		return "Authorization: Bearer " + t
	}
	return ""
}
GO

# --- Cobra command (if repo uses cobra) ---
cat > cmd/restless/validate_cmd.go <<'GO'
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bspippi1337/restless/internal/validate"
)

// If your project already uses Cobra, you can ignore this file and wire validate.Run() into your command tree.
// This file provides a fallback "subcommand" style entry for repos that use basic flag parsing.
//
// Usage:
//   restless validate --spec openapi.yaml --base https://api.example.com
func validateMain(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	spec := fs.String("spec", "", "Path to OpenAPI spec (yaml/json)")
	base := fs.String("base", "", "Base URL (e.g. https://api.example.com)")
	timeout := fs.Duration("timeout", 15*time.Second, "HTTP timeout")
	strict := fs.Bool("strict", false, "Strict mode (fail on any unexpected status codes)")
	jsonOut := fs.Bool("json", false, "Machine-readable JSON output")
	reportDir := fs.String("report", "", "Optional report directory to write validate.json")

	auth := fs.String("auth", "", `Optional header "Key: Value" (or set RESTLESS_AUTH / RESTLESS_TOKEN)`)

	if err := fs.Parse(args); err != nil {
		return 2
	}

	a := *auth
	if a == "" {
		a = validate.AuthHeaderFromEnv()
	}

	rep, err := validate.Run(context.Background(), validate.Options{
		SpecPath:   *spec,
		BaseURL:    *base,
		Timeout:    *timeout,
		StrictLive: *strict,
		AuthHeader: a,
		JSON:       *jsonOut,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "validate error:", err)
		return 2
	}

	if *jsonOut {
		_ = validate.PrintJSON(rep, os.Stdout)
	} else {
		validate.PrintHuman(rep, os.Stdout)
	}

	if *reportDir != "" {
		_ = os.MkdirAll(*reportDir, 0o755)
		f, ferr := os.Create(*reportDir + "/validate.json")
		if ferr == nil {
			defer f.Close()
			_ = validate.PrintJSON(rep, f)
		}
	}

	if rep.OK {
		return 0
	}
	return 1
}
GO

# --- Wire into main: detect simple subcommand router ---
MAIN_FILE="cmd/restless/main.go"
if [ -f "$MAIN_FILE" ]; then
  if grep -q 'sp13.*cobra\|spf13/cobra' "$MAIN_FILE" || grep -q 'spf13/cobra' go.mod 2>/dev/null; then
    echo "==> Detected cobra. Adding a patch note instead of auto-edit (safer)."
    cat > cmd/restless/VALIDATE_WIRING_NOTE.md <<'MD'
# Wiring `validate` into Cobra

This repo appears to use Cobra.

Wire in `internal/validate` by adding a `validate` subcommand that calls:

- `validate.Run(ctx, validate.Options{...})`
- print via `validate.PrintHuman` or `validate.PrintJSON`
- exit codes: 0 OK, 1 drift, 2 runtime error

Files added:
- internal/validate/validate.go
- cmd/restless/validate_cmd.go (fallback implementation, can be ignored if using Cobra)

Suggested Cobra handler signature:
- `RunE: func(cmd *cobra.Command, args []string) error { ... }`
MD
  else
    echo "==> Attempting to auto-wire validate into cmd/restless/main.go (simple subcommand style)..."
    # Insert a minimal dispatch: if os.Args[1]=="validate" call validateMain(os.Args[2:])
    # Only if not already present
    if ! grep -q 'validateMain' "$MAIN_FILE"; then
      perl -0777 -i -pe 's/package main\s*/package main\n\n/; s/(func main\(\)\s*\{\s*)/$1\n\tif len(os.Args) > 1 && os.Args[1] == "validate" {\n\t\tos.Exit(validateMain(os.Args[2:]))\n\t}\n/;' "$MAIN_FILE" || true
      # ensure os imported
      if ! grep -q '"os"' "$MAIN_FILE"; then
        perl -0777 -i -pe 's/import\s*\(\s*/import (\n\t"os"\n/;' "$MAIN_FILE" || true
      fi
    fi
  fi
else
  echo "==> NOTE: cmd/restless/main.go not found. You must wire validateMain() manually."
fi

echo "==> Formatting..."
gofmt -w internal/validate/validate.go cmd/restless/validate_cmd.go >/dev/null || true
[ -f cmd/restless/main.go ] && gofmt -w cmd/restless/main.go >/dev/null || true

echo "==> Building..."
go build ./cmd/restless

echo "==> Committing..."
git add -A
git commit -m "feat(validate): add OpenAPI vs live drift gatekeeper" || true

echo
echo "✅ Done."
echo
echo "Try:"
echo "  restless validate --spec openapi.yaml --base https://api.example.com"
echo
echo "Optional auth:"
echo '  export RESTLESS_TOKEN="xxx"'
echo '  # or'
echo '  export RESTLESS_AUTH="Authorization: Bearer xxx"'
