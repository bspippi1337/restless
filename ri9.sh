bash -c 'set -euo pipefail

# =========================================
# Restless: Build + Ship (APT/Brew) + Smart Client
# - Adds a fast stdlib-only smart client as cmd/restless-smart (no bubbletea/clipboard/tty)
# - Adds GitHub Actions:
#     1) release.yml (build+release assets)
#     2) apt-pages.yml (flat APT repo on gh-pages, no reprepro)
#     3) brew-tap.yml (auto-update Homebrew tap formula repo)
# - Adds scripts/ship-all.sh (local one-shot: test, build, tag, release, apt pages, brew tap)
# =========================================

mkdir -p .github/workflows scripts cmd/restless-smart internal/smartclient docs

# ---------------------------
# 1) Smart client (stdlib-only)
# ---------------------------
cat > cmd/restless-smart/main.go <<'\''GO'\''
//go:build !wasm

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Finding struct {
	URL           string            `json:"url"`
	ResolvedURL   string            `json:"resolved_url,omitempty"`
	Status        int               `json:"status"`
	Methods       []string          `json:"methods,omitempty"`
	ContentType   string            `json:"content_type,omitempty"`
	Server        string            `json:"server,omitempty"`
	AuthHints     []string          `json:"auth_hints,omitempty"`
	RateLimit     map[string]string `json:"rate_limit,omitempty"`
	OpenAPI       string            `json:"openapi,omitempty"`
	OpenAPIPaths  int               `json:"openapi_paths,omitempty"`
	Suggestions   []string          `json:"suggestions,omitempty"`
	Warnings      []string          `json:"warnings,omitempty"`
	DiscoveredAt  string            `json:"discovered_at"`
	ElapsedMillis int64             `json:"elapsed_ms"`
}

func main() {
	var (
		timeout   = flag.Duration("timeout", 6*time.Second, "request timeout")
		ua        = flag.String("ua", "restless-smart/1.0", "User-Agent")
		export    = flag.String("export", "", "export format: json|md|curl|har (writes to stdout)")
		method    = flag.String("method", "GET", "method for curl/har export")
		body      = flag.String("body", "", "body for curl/har export")
		headerKVs = flag.String("H", "", "extra headers: \"K: V\\nK2: V2\"")
	)
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: restless-smart [--export fmt] <url>\nExample: restless-smart https://api.example.com\nExport:  restless-smart --export=curl --method=POST --body='\''{\"x\":1}'\'' https://api.example.com")
		os.Exit(2)
	}
	raw := args[0]
	u, err := normalizeURL(raw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid url:", err)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	start := time.Now()
	f, err := smartProfile(ctx, u, *ua, *headerKVs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "smart profile failed:", err)
		os.Exit(1)
	}
	f.DiscoveredAt = time.Now().UTC().Format(time.RFC3339)
	f.ElapsedMillis = time.Since(start).Milliseconds()

	switch strings.ToLower(strings.TrimSpace(*export)) {
	case "":
		printHuman(f)
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(f)
	case "md", "markdown":
		fmt.Print(exportMarkdown(f))
	case "curl":
		fmt.Print(exportCurl(*method, u, *headerKVs, *body))
	case "har":
		fmt.Print(exportHAR(*method, u, *headerKVs, *body))
	default:
		fmt.Fprintln(os.Stderr, "unknown export format:", *export)
		os.Exit(2)
	}
}

func normalizeURL(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("empty")
	}
	if !strings.Contains(s, "://") {
		// default to https, fall back to http if needed during probing
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", errors.New("missing host")
	}
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String(), nil
}

func smartProfile(ctx context.Context, target, ua, headerKVs string) (*Finding, error) {
	f := &Finding{
		URL:         target,
		RateLimit:   map[string]string{},
		AuthHints:   []string{},
		Suggestions: []string{},
		Warnings:    []string{},
	}

	// We try HTTPS first (as normalized), then HTTP if TLS fails.
	tries := []string{target}
	if strings.HasPrefix(target, "https://") {
		tries = append(tries, "http://"+strings.TrimPrefix(target, "https://"))
	}

	var lastErr error
	for _, u := range tries {
		res, err := doProbe(ctx, u, ua, headerKVs)
		if err != nil {
			lastErr = err
			continue
		}
		*f = *res
		break
	}
	if f.ResolvedURL == "" && lastErr != nil {
		return nil, lastErr
	}

	// OpenAPI discovery (common paths)
	openapiURL, paths, warn := tryOpenAPI(ctx, f.ResolvedURL, ua, headerKVs)
	if warn != "" {
		f.Warnings = append(f.Warnings, warn)
	}
	if openapiURL != "" {
		f.OpenAPI = openapiURL
		f.OpenAPIPaths = paths
		f.Suggestions = append(f.Suggestions, "openapi: detected, consider generating typed clients or docs")
	}

	// Suggestions based on headers/status
	if f.ContentType == "application/json" {
		f.Suggestions = append(f.Suggestions, "json: good candidate for export formats (json/md/har/curl)")
	}
	if len(f.AuthHints) > 0 {
		f.Suggestions = append(f.Suggestions, "auth: add Authorization header, or token via env var")
	}
	if len(f.Methods) == 0 {
		f.Suggestions = append(f.Suggestions, "methods: try OPTIONS or consult docs")
	} else {
		f.Suggestions = append(f.Suggestions, "methods: "+strings.Join(f.Methods, ", "))
	}
	if f.Status >= 400 {
		f.Suggestions = append(f.Suggestions, "status: non-2xx, verify base URL/path and auth")
	}

	return f, nil
}

func doProbe(ctx context.Context, target, ua, headerKVs string) (*Finding, error) {
	client := &http.Client{
		Timeout: 0, // governed by ctx
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 8 {
				return errors.New("too many redirects")
			}
			return nil
		},
	}

	// Prefer HEAD; if not allowed, fall back to GET.
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, target, nil)
	req.Header.Set("User-Agent", ua)
	applyExtraHeaders(req, headerKVs)

	resp, err := client.Do(req)
	if err != nil {
		// HEAD sometimes rejected by intermediaries; try GET as fallback for reachability.
		req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		req2.Header.Set("User-Agent", ua)
		applyExtraHeaders(req2, headerKVs)
		resp2, err2 := client.Do(req2)
		if err2 != nil {
			return nil, err
		}
		resp = resp2
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))

	f := &Finding{
		URL:         target,
		ResolvedURL: resp.Request.URL.String(),
		Status:      resp.StatusCode,
		RateLimit:   map[string]string{},
		AuthHints:   []string{},
		Suggestions: []string{},
		Warnings:    []string{},
	}

	// Content-Type
	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		ct = strings.Split(ct, ";")[0]
		ct = strings.TrimSpace(ct)
	}
	f.ContentType = ct

	// Server
	f.Server = resp.Header.Get("Server")

	// Allow / Access-Control-Allow-Methods (common)
	methods := parseMethods(resp.Header.Get("Allow"))
	if len(methods) == 0 {
		methods = parseMethods(resp.Header.Get("Access-Control-Allow-Methods"))
	}
	f.Methods = methods

	// Auth hints
	if www := resp.Header.Values("WWW-Authenticate"); len(www) > 0 {
		for _, v := range www {
			f.AuthHints = append(f.AuthHints, strings.TrimSpace(v))
		}
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		f.AuthHints = append(f.AuthHints, "status suggests auth required")
	}

	// Rate limit headers (best-effort)
	for _, k := range []string{
		"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset",
		"RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset",
		"Retry-After",
	} {
		if v := resp.Header.Get(k); v != "" {
			f.RateLimit[k] = v
		}
	}

	return f, nil
}

func tryOpenAPI(ctx context.Context, base, ua, headerKVs string) (found string, paths int, warning string) {
	candidates := []string{
		"/openapi.json", "/openapi.yaml", "/swagger.json", "/swagger.yaml",
		"/api-docs", "/api/docs", "/v3/api-docs", "/v2/api-docs",
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return "", 0, ""
	}

	client := &http.Client{Timeout: 0}
	for _, p := range candidates {
		u := *baseURL
		u.Path = joinPath(u.Path, p)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		req.Header.Set("User-Agent", ua)
		applyExtraHeaders(req, headerKVs)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			continue
		}
		ct := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
		ct = strings.TrimSpace(ct)

		// If JSON, try count paths
		if strings.Contains(ct, "json") || (len(b) > 0 && (b[0] == '{' || b[0] == '[')) {
			var obj map[string]any
			if json.Unmarshal(b, &obj) == nil {
				if pv, ok := obj["paths"].(map[string]any); ok {
					return u.String(), len(pv), ""
				}
				// Some servers wrap docs differently
				return u.String(), 0, "openapi: detected but could not count paths"
			}
		}
		// YAML: do not parse without deps, but we can still claim detection
		if strings.Contains(ct, "yaml") || strings.Contains(ct, "yml") || bytes.Contains(b, []byte("openapi:")) || bytes.Contains(b, []byte("swagger:")) {
			return u.String(), 0, "openapi: yaml detected (path count unavailable without yaml parser)"
		}
	}
	return "", 0, ""
}

func joinPath(a, b string) string {
	// ensure single slash join
	if a == "" {
		a = "/"
	}
	if !strings.HasSuffix(a, "/") {
		// keep base path as-is for typical API roots; we only append known doc paths at root
	}
	if strings.HasPrefix(b, "/") {
		return b
	}
	return "/" + b
}

func applyExtraHeaders(req *http.Request, headerKVs string) {
	headerKVs = strings.TrimSpace(headerKVs)
	if headerKVs == "" {
		return
	}
	sc := bufio.NewScanner(strings.NewReader(headerKVs))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		i := strings.Index(line, ":")
		if i <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		if k != "" {
			req.Header.Add(k, v)
		}
	}
}

func parseMethods(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		m := strings.ToUpper(strings.TrimSpace(p))
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	return out
}

func printHuman(f *Finding) {
	fmt.Println("🧠 Restless Smart Client")
	fmt.Println("URL:       ", f.URL)
	if f.ResolvedURL != "" && f.ResolvedURL != f.URL {
		fmt.Println("Resolved:  ", f.ResolvedURL)
	}
	fmt.Println("Status:    ", f.Status)
	if f.ContentType != "" {
		fmt.Println("Type:      ", f.ContentType)
	}
	if f.Server != "" {
		fmt.Println("Server:    ", f.Server)
	}
	if len(f.Methods) > 0 {
		fmt.Println("Methods:   ", strings.Join(f.Methods, ", "))
	}
	if len(f.AuthHints) > 0 {
		fmt.Println("Auth:      ", strings.Join(f.AuthHints, " | "))
	}
	if len(f.RateLimit) > 0 {
		fmt.Println("RateLimit: ")
		for k, v := range f.RateLimit {
			fmt.Printf("  - %s: %s\n", k, v)
		}
	}
	if f.OpenAPI != "" {
		if f.OpenAPIPaths > 0 {
			fmt.Printf("OpenAPI:   %s (%d paths)\n", f.OpenAPI, f.OpenAPIPaths)
		} else {
			fmt.Printf("OpenAPI:   %s\n", f.OpenAPI)
		}
	}
	if len(f.Warnings) > 0 {
		fmt.Println("Warnings:  ", strings.Join(f.Warnings, " | "))
	}
	if len(f.Suggestions) > 0 {
		fmt.Println("\nSuggested next steps:")
		for _, s := range f.Suggestions {
			fmt.Println("  •", s)
		}
	}
	fmt.Printf("\nElapsed:   %dms\n", f.ElapsedMillis)
}

func exportMarkdown(f *Finding) string {
	var b strings.Builder
	b.WriteString("# Restless Smart Profile\n\n")
	b.WriteString(fmt.Sprintf("- URL: `%s`\n", f.URL))
	if f.ResolvedURL != "" && f.ResolvedURL != f.URL {
		b.WriteString(fmt.Sprintf("- Resolved: `%s`\n", f.ResolvedURL))
	}
	b.WriteString(fmt.Sprintf("- Status: `%d`\n", f.Status))
	if f.ContentType != "" {
		b.WriteString(fmt.Sprintf("- Content-Type: `%s`\n", f.ContentType))
	}
	if len(f.Methods) > 0 {
		b.WriteString(fmt.Sprintf("- Methods: `%s`\n", strings.Join(f.Methods, ", ")))
	}
	if f.OpenAPI != "" {
		if f.OpenAPIPaths > 0 {
			b.WriteString(fmt.Sprintf("- OpenAPI: `%s` (%d paths)\n", f.OpenAPI, f.OpenAPIPaths))
		} else {
			b.WriteString(fmt.Sprintf("- OpenAPI: `%s`\n", f.OpenAPI))
		}
	}
	if len(f.AuthHints) > 0 {
		b.WriteString("\n## Auth hints\n")
		for _, h := range f.AuthHints {
			b.WriteString("- " + h + "\n")
		}
	}
	if len(f.RateLimit) > 0 {
		b.WriteString("\n## Rate limit\n")
		for k, v := range f.RateLimit {
			b.WriteString(fmt.Sprintf("- `%s`: `%s`\n", k, v))
		}
	}
	if len(f.Suggestions) > 0 {
		b.WriteString("\n## Suggestions\n")
		for _, s := range f.Suggestions {
			b.WriteString("- " + s + "\n")
		}
	}
	return b.String()
}

func exportCurl(method, target, headerKVs, body string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "GET"
	}
	var b strings.Builder
	b.WriteString("curl -sS ")
	b.WriteString("-X " + method + " ")
	if strings.TrimSpace(headerKVs) != "" {
		sc := bufio.NewScanner(strings.NewReader(headerKVs))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || !strings.Contains(line, ":") {
				continue
			}
			b.WriteString("-H ")
			b.WriteString(shellQuote(line))
			b.WriteString(" ")
		}
	}
	if strings.TrimSpace(body) != "" {
		b.WriteString("-d ")
		b.WriteString(shellQuote(body))
		b.WriteString(" ")
	}
	b.WriteString(shellQuote(target))
	b.WriteString("\n")
	return b.String()
}

func exportHAR(method, target, headerKVs, body string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "GET"
	}
	type H struct{ Name, Value string }
	hdrs := []H{}
	if strings.TrimSpace(headerKVs) != "" {
		sc := bufio.NewScanner(strings.NewReader(headerKVs))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			i := strings.Index(line, ":")
			if i <= 0 {
				continue
			}
			hdrs = append(hdrs, H{Name: strings.TrimSpace(line[:i]), Value: strings.TrimSpace(line[i+1:])})
		}
	}
	out := map[string]any{
		"log": map[string]any{
			"version": "1.2",
			"creator": map[string]any{"name": "restless-smart", "version": "1.0"},
			"entries": []any{
				map[string]any{
					"startedDateTime": time.Now().UTC().Format(time.RFC3339),
					"request": map[string]any{
						"method":  method,
						"url":     target,
						"headers": hdrs,
						"postData": func() any {
							if strings.TrimSpace(body) == "" {
								return nil
							}
							return map[string]any{"mimeType": "application/json", "text": body}
						}(),
					},
				},
			},
		},
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b) + "\n"
}

func shellQuote(s string) string {
	// POSIX-ish single-quote safe
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n'\"\\$`") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
GO

# Optional: small internal note for docs
cat > internal/smartclient/README.md <<'\''EOF'\''
Smart Client is shipped as a separate binary: cmd/restless-smart

Why separate?
- Keeps core restless binary untouched
- Avoids TUI/clipboard/tty deps for browser/CI builds
- Enables fast iteration on profiling logic

Usage:
  restless-smart https://api.example.com
  restless-smart --export=json https://api.example.com
  restless-smart --export=curl --method=POST --body='{"x":1}' https://api.example.com
EOF

# ---------------------------
# 2) Release workflow (build+assets)
# ---------------------------
cat > .github/workflows/release.yml <<'\''YAML'\''
name: Build + Release Assets

on:
  push:
    tags:
      - "v*"
  workflow_dispatch: {}

permissions:
  contents: write

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin]
        goarch: [amd64]
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "stable"

      - name: Test (pure go)
        run: |
          CGO_ENABLED=0 go test ./...

      - name: Build binaries
        run: |
          mkdir -p dist
          CGO_ENABLED=0 GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o dist/restless_${{ matrix.goos }}_${{ matrix.goarch }} ./cmd/restless
          CGO_ENABLED=0 GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o dist/restless-smart_${{ matrix.goos }}_${{ matrix.goarch }} ./cmd/restless-smart

      - name: Package tar.gz
        run: |
          cd dist
          tar -czf restless_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz restless_${{ matrix.goos }}_${{ matrix.goarch }} restless-smart_${{ matrix.goos }}_${{ matrix.goarch }}
          sha256sum restless_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz > restless_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz.sha256

      - name: Upload to Release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          TAG="${GITHUB_REF_NAME}"
          gh release view "$TAG" >/dev/null 2>&1 || gh release create "$TAG" -t "$TAG" -n "Release $TAG"
          gh release upload "$TAG" dist/restless_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz --clobber
          gh release upload "$TAG" dist/restless_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz.sha256 --clobber
YAML

# ---------------------------
# 3) APT Pages workflow (flat repo, no reprepro)
# ---------------------------
cat > .github/workflows/apt-pages.yml <<'\''YAML'\''
name: Publish APT (flat) to GitHub Pages

on:
  push:
    tags:
      - "v*"
  workflow_dispatch: {}

permissions:
  contents: write

jobs:
  apt:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "stable"

      - name: Install tooling
        run: |
          sudo apt-get update
          sudo apt-get install -y dpkg-dev

      - name: Build linux binaries
        run: |
          mkdir -p dist
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/restless ./cmd/restless
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/restless-smart ./cmd/restless-smart
          chmod +x dist/restless dist/restless-smart

      - name: Build .deb (includes both binaries)
        run: |
          TAG="${GITHUB_REF_NAME}"
          VERSION="${TAG#v}"
          mkdir -p pkg/DEBIAN pkg/usr/bin
          install -m 0755 dist/restless pkg/usr/bin/restless
          install -m 0755 dist/restless-smart pkg/usr/bin/restless-smart
          cat > pkg/DEBIAN/control <<CTL
Package: restless
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: amd64
Maintainer: bspippi1337 <noreply@github.com>
Description: Restless universal API client + restless-smart profiler
CTL
          dpkg-deb --build pkg dist/restless_${VERSION}_amd64.deb
          rm -rf pkg

      - name: Create flat APT repo
        run: |
          mkdir -p apt-repo
          cp dist/*.deb apt-repo/
          cd apt-repo
          dpkg-scanpackages . /dev/null > Packages
          gzip -k -f Packages
          cat > Release <<REL
Origin: restless
Label: restless APT
Suite: stable
Codename: stable
Architectures: amd64
Components: main
Description: APT repo for restless
REL
          cd ..
          touch apt-repo/.nojekyll

      - name: Publish gh-pages
        run: |
          git config user.name github-actions
          git config user.email github-actions@github.com
          git fetch origin gh-pages || true
          git checkout gh-pages || git checkout --orphan gh-pages
          rm -rf *
          cp -a apt-repo/. .
          git add -A
          git commit -m "APT publish ${GITHUB_REF_NAME}" || true
  
