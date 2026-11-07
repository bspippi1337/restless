#!/usr/bin/env bash
set -euo pipefail

# Restless v2 "Over The Top" script
# Safe-by-default: makes a work branch, snapshots, applies staged commits, runs fmt/tidy/test.
# Idempotent-ish: re-running should mostly be safe (may no-op if already applied).
#
# Usage:
#   bash scripts/v2_over_the_top.sh
#
# Optional env:
#   WORK_BRANCH=work/v2-over-the-top
#   SKIP_SNAPSHOT=1
#   NO_COMMIT=1

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WORK_BRANCH="${WORK_BRANCH:-work/v2-over-the-top}"
SKIP_SNAPSHOT="${SKIP_SNAPSHOT:-0}"
NO_COMMIT="${NO_COMMIT:-0}"

say() { printf "%s\n" "$*"; }
die() { printf "ERROR: %s\n" "$*" >&2; exit 1; }

have() { command -v "$1" >/dev/null 2>&1; }

git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "Not a git repo."
have go || die "Go is not installed / not in PATH."

stage_commit() {
  local msg="$1"
  if [[ "$NO_COMMIT" == "1" ]]; then
    say "==> NO_COMMIT=1: skipping commit: $msg"
    return 0
  fi
  if git diff --quiet && git diff --cached --quiet; then
    say "==> No changes to commit for: $msg"
    return 0
  fi
  git add -A
  git commit -m "$msg" >/dev/null 2>&1 || true
  say "==> committed: $msg"
}

snapshot() {
  [[ "$SKIP_SNAPSHOT" == "1" ]] && { say "==> SKIP_SNAPSHOT=1: skipping snapshot"; return 0; }

  mkdir -p .restless_snapshots
  local ts
  ts="$(date +%Y%m%d-%H%M%S)"
  local snap_branch="safety/${ts}"

  if git show-ref --verify --quiet "refs/heads/${snap_branch}"; then
    say "==> safety branch already exists: ${snap_branch}"
  else
    git switch -c "${snap_branch}" >/dev/null 2>&1 || git checkout -b "${snap_branch}"
    say "==> created safety branch: ${snap_branch}"
  fi

  local zip_path=".restless_snapshots/repo-${ts}.zip"
  say "==> snapshot zip: ${zip_path}"
  (
    shopt -s dotglob
    zip -qr "${zip_path}" . \
      -x ".git/*" \
      -x ".restless_snapshots/*" \
      -x "dist/*" \
      -x "bin/*" \
      -x "**/.DS_Store" \
      -x "**/node_modules/*" \
      -x "**/.idea/*" \
      -x "**/.vscode/*"
  )
  say "==> snapshot done"

  # Return to original branch (work branch will be created next)
  git switch - >/dev/null 2>&1 || true
}

switch_work_branch() {
  if git show-ref --verify --quiet "refs/heads/${WORK_BRANCH}"; then
    git switch "${WORK_BRANCH}" >/dev/null 2>&1 || git checkout "${WORK_BRANCH}"
  else
    git switch -c "${WORK_BRANCH}" >/dev/null 2>&1 || git checkout -b "${WORK_BRANCH}"
  fi
  say "==> on branch: ${WORK_BRANCH}"
}

fmt_tidy_test() {
  say "==> gofmt"
  gofmt -w . >/dev/null 2>&1 || true

  say "==> go mod tidy"
  go mod tidy >/dev/null 2>&1 || true

  say "==> go test ./..."
  go test ./... >/dev/null
}

mkdir -p scripts

# ---------------------------
# 0) Safety + work branch
# ---------------------------
snapshot
switch_work_branch

# ---------------------------
# 1) v2 directory skeleton
# ---------------------------
say "==> Phase 1: v2 structure"
mkdir -p internal/core/{app,config,httpx,engine,store,types,logx}
mkdir -p internal/modules/{openapi,session,bench,export,tui,gui}
mkdir -p internal/ui/{cli,tui,gui}
mkdir -p docs examples .restless

cat > docs/ARCHITECTURE.md <<'EOF'
# Restless v2 Architecture

## Goals
- Single binary.
- Modular codebase (compile-time modules).
- Small, stable core.
- Modules can evolve fast without breaking the core.

## Rules
- `internal/core` MUST NOT import `internal/modules`.
- `internal/modules` MAY import `internal/core`.
- UI layers (cli/tui/gui) depend on core + modules, never the other way around.

## Layout
- `internal/core/*`  : stable foundation (config, transport, engine, store, shared types)
- `internal/modules/*`: feature modules (openapi, sessions, bench, export, ui modules)
- `internal/ui/*`    : renderers and interaction layers
EOF

stage_commit "chore: add v2 architecture skeleton"

# ---------------------------
# 2) Core types + engine contract
# ---------------------------
say "==> Phase 2: core engine contract + runner"

cat > internal/core/types/types.go <<'EOF'
package types

import "net/http"

// Request is the normalized request model used across CLI/TUI/GUI/modules.
type Request struct {
	Method  string
	URL     string
	Headers http.Header
	Body    []byte
}

// Response is the normalized response model used across CLI/TUI/GUI/modules.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	DurationMs int64
}
EOF

cat > internal/core/engine/contract.go <<'EOF'
package engine

import (
	"context"

	"github.com/bspippi1337/restless/internal/core/types"
)

// Runner executes a request and returns a normalized response.
// This is the stable "spine" of Restless v2.
type Runner interface {
	Run(ctx context.Context, req types.Request) (types.Response, error)
}
EOF

cat > internal/core/httpx/httpx.go <<'EOF'
package httpx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/bspippi1337/restless/internal/core/types"
)

// DefaultClient returns a reasonably safe default HTTP client.
func DefaultClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// Do executes an HTTP request using the provided client.
func Do(ctx context.Context, c *http.Client, req types.Request) (types.Response, error) {
	start := time.Now()

	hreq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return types.Response{}, err
	}
	if req.Headers != nil {
		hreq.Header = req.Headers.Clone()
	}

	resp, err := c.Do(hreq)
	if err != nil {
		return types.Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.Response{}, err
	}

	dur := time.Since(start).Milliseconds()
	return types.Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
		DurationMs: dur,
	}, nil
}
EOF

cat > internal/core/engine/http_runner.go <<'EOF'
package engine

import (
	"context"
	"net/http"

	"github.com/bspippi1337/restless/internal/core/httpx"
	"github.com/bspippi1337/restless/internal/core/types"
)

type HTTPRunner struct {
	Client *http.Client
}

func NewHTTPRunner(c *http.Client) *HTTPRunner {
	if c == nil {
		c = httpx.DefaultClient()
	}
	return &HTTPRunner{Client: c}
}

func (r *HTTPRunner) Run(ctx context.Context, req types.Request) (types.Response, error) {
	return httpx.Do(ctx, r.Client, req)
}
EOF

cat > internal/core/logx/logx.go <<'EOF'
package logx

import (
	"fmt"
	"log"
	"os"
)

type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
)

type Logger struct {
	l     *log.Logger
	level Level
}

func New(level Level) *Logger {
	return &Logger{
		l:     log.New(os.Stderr, "", log.LstdFlags),
		level: level,
	}
}

func (lg *Logger) Printf(level Level, format string, args ...any) {
	// Simple level gate (debug is most verbose)
	if lg.level != Debug && level == Debug {
		return
	}
	lg.l.Printf("[%s] %s", level, fmt.Sprintf(format, args...))
}
EOF

stage_commit "feat(core): introduce v2 engine contract + http runner"

# ---------------------------
# 3) Module registry (compile-time, plugin-ready)
# ---------------------------
say "==> Phase 3: compile-time module registry"

cat > internal/core/app/registry.go <<'EOF'
package app

import (
	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/bspippi1337/restless/internal/core/logx"
)

// Module is a compile-time feature module.
// Future plugin runtimes can implement the same interface (WASM, etc).
type Module interface {
	Name() string
	Register(r *Registry) error
}

// Registry is the wiring surface between core and modules.
// Keep it stable and small.
type Registry struct {
	Log    *logx.Logger
	Runner engine.Runner

	// Hooks
	RequestMutators  []func(*RequestContext) error
	ResponseMutators []func(*ResponseContext) error
}

type RequestContext struct {
	// For future: env/profile/session vars etc
	Method string
	URL    string
	Body   []byte
	Header map[string][]string
}

type ResponseContext struct {
	StatusCode int
	Body       []byte
	Header     map[string][]string
}

func NewRegistry(lg *logx.Logger, runner engine.Runner) *Registry {
	if lg == nil {
		lg = logx.New(logx.Info)
	}
	return &Registry{
		Log:    lg,
		Runner: runner,
	}
}
EOF

cat > internal/core/app/app.go <<'EOF'
package app

import (
	"context"
	"net/http"

	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/bspippi1337/restless/internal/core/logx"
	"github.com/bspippi1337/restless/internal/core/types"
)

type App struct {
	reg *Registry
}

func New(mods []Module) (*App, error) {
	lg := logx.New(logx.Info)
	runner := engine.NewHTTPRunner(nil)
	reg := NewRegistry(lg, runner)

	for _, m := range mods {
		reg.Log.Printf(logx.Info, "registering module: %s", m.Name())
		if err := m.Register(reg); err != nil {
			return nil, err
		}
	}

	return &App{reg: reg}, nil
}

func (a *App) RunOnce(ctx context.Context, req types.Request) (types.Response, error) {
	// Apply request mutators (future: template vars, auth, etc)
	rc := &RequestContext{
		Method: req.Method,
		URL:    req.URL,
		Body:   req.Body,
		Header: map[string][]string{},
	}
	if req.Headers != nil {
		for k, vv := range req.Headers {
			rc.Header[k] = append([]string{}, vv...)
		}
	}
	for _, fn := range a.reg.RequestMutators {
		if err := fn(rc); err != nil {
			return types.Response{}, err
		}
	}

	// Build final request
	h := http.Header{}
	for k, vv := range rc.Header {
		for _, v := range vv {
			h.Add(k, v)
		}
	}
	finalReq := types.Request{
		Method:  rc.Method,
		URL:     rc.URL,
		Headers: h,
		Body:    rc.Body,
	}

	resp, err := a.reg.Runner.Run(ctx, finalReq)
	if err != nil {
		return types.Response{}, err
	}

	// Apply response mutators
	rsc := &ResponseContext{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Header:     map[string][]string{},
	}
	for k, vv := range resp.Headers {
		rsc.Header[k] = append([]string{}, vv...)
	}
	for _, fn := range a.reg.ResponseMutators {
		if err := fn(rsc); err != nil {
			return types.Response{}, err
		}
	}

	// Write back
	outHeaders := http.Header{}
	for k, vv := range rsc.Header {
		for _, v := range vv {
			outHeaders.Add(k, v)
		}
	}
	return types.Response{
		StatusCode: rsc.StatusCode,
		Headers:    outHeaders,
		Body:       rsc.Body,
		DurationMs: resp.DurationMs,
	}, nil
}
EOF

stage_commit "feat(core): add module registry and app wiring surface"

# ---------------------------
# 4) Sessions module (vars + extractors skeleton)
# ---------------------------
say "==> Phase 4: sessions module v1 skeleton"

cat > internal/modules/session/module.go <<'EOF'
package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/bspippi1337/restless/internal/core/app"
)

// Module provides session vars + templating hooks.
type Module struct {
	vars map[string]string
}

func New() *Module {
	return &Module{vars: map[string]string{}}
}

func (m *Module) Name() string { return "session" }

func (m *Module) Register(r *app.Registry) error {
	// Request templating: replace {{var}} in URL and body
	r.RequestMutators = append(r.RequestMutators, func(rc *app.RequestContext) error {
		rc.URL = applyTemplates(rc.URL, m.vars)
		if len(rc.Body) > 0 {
			rc.Body = []byte(applyTemplates(string(rc.Body), m.vars))
		}
		// headers
		for k, vv := range rc.Header {
			for i := range vv {
				vv[i] = applyTemplates(vv[i], m.vars)
			}
			rc.Header[k] = vv
		}
		return nil
	})
	return nil
}

// Set sets a session var (string).
func (m *Module) Set(key, value string) {
	if key == "" {
		return
	}
	m.vars[key] = value
}

// ExtractJSON extracts a value from JSON response by a simple dot path (v1).
func (m *Module) ExtractJSON(dotPath string, body []byte) (string, error) {
	if dotPath == "" {
		return "", errors.New("empty path")
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return "", err
	}
	parts := bytes.Split([]byte(dotPath), []byte("."))
	cur := v
	for _, p := range parts {
		key := string(p)
		obj, ok := cur.(map[string]any)
		if !ok {
			return "", errors.New("path not found")
		}
		cur, ok = obj[key]
		if !ok {
			return "", errors.New("path not found")
		}
	}
	switch t := cur.(type) {
	case string:
		return t, nil
	default:
		b, _ := json.Marshal(t)
		return string(b), nil
	}
}

// ExtractRegex extracts first capture group from body text.
func (m *Module) ExtractRegex(pattern string, body []byte) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	mm := re.FindSubmatch(body)
	if len(mm) < 2 {
		return "", errors.New("no match")
	}
	return string(mm[1]), nil
}

var tmplRe = regexp.MustCompile(`\{\{([a-zA-Z0-9_.-]+)\}\}`)

func applyTemplates(s string, vars map[string]string) string {
	return tmplRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := tmplRe.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		if v, ok := vars[sub[1]]; ok {
			return v
		}
		return m
	})
}
EOF

stage_commit "feat(modules): add sessions module v1 skeleton (vars + extractors)"

# ---------------------------
# 5) OpenAPI module (import/cache/list skeleton)
# ---------------------------
say "==> Phase 5: openapi module v1 skeleton"

cat > internal/modules/openapi/module.go <<'EOF'
package openapi

import "github.com/bspippi1337/restless/internal/core/app"

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Name() string { return "openapi" }

func (m *Module) Register(r *app.Registry) error {
	// v1 skeleton: no hooks yet (CLI commands will use functions directly).
	return nil
}
EOF

cat > internal/modules/openapi/store.go <<'EOF'
package openapi

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

type SpecIndex struct {
	ID        string `json:"id"`
	Source    string `json:"source"` // URL or file path
	Imported  int64  `json:"imported_unix"`
	Title     string `json:"title"`
	Version   string `json:"version"`
	RawPath   string `json:"raw_path"` // stored raw file
	IndexPath string `json:"index_path"`
}

func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".restless", "openapi"), nil
}

func idFor(source string) string {
	h := sha1.Sum([]byte(source))
	return hex.EncodeToString(h[:])
}

func SaveIndex(idx SpecIndex) error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(dir, idx.ID+".json")
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

func LoadIndex(id string) (SpecIndex, error) {
	dir, err := cacheDir()
	if err != nil {
		return SpecIndex{}, err
	}
	p := filepath.Join(dir, id+".json")
	b, err := os.ReadFile(p)
	if err != nil {
		return SpecIndex{}, err
	}
	var idx SpecIndex
	if err := json.Unmarshal(b, &idx); err != nil {
		return SpecIndex{}, err
	}
	return idx, nil
}
EOF

cat > internal/modules/openapi/importer.go <<'EOF'
package openapi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Import fetches from URL or reads from file, then stores raw content in ~/.restless/openapi/<id>.yaml|json
// v1 skeleton: we don't parse endpoints yet; we just cache and create an index entry.
func Import(source string) (SpecIndex, error) {
	if source == "" {
		return SpecIndex{}, errors.New("empty source")
	}

	id := idFor(source)
	dir, err := cacheDir()
	if err != nil {
		return SpecIndex{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return SpecIndex{}, err
	}

	rawPath := filepath.Join(dir, id+".raw")
	var data []byte

	if looksLikeURL(source) {
		resp, err := http.Get(source) //nolint:gosec
		if err != nil {
			return SpecIndex{}, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return SpecIndex{}, err
		}
		data = b
	} else {
		b, err := os.ReadFile(source)
		if err != nil {
			return SpecIndex{}, err
		}
		data = b
	}

	if err := os.WriteFile(rawPath, data, 0o644); err != nil {
		return SpecIndex{}, err
	}

	idx := SpecIndex{
		ID:       id,
		Source:   source,
		Imported: time.Now().Unix(),
		Title:    "",
		Version:  "",
		RawPath:  rawPath,
	}

	if err := SaveIndex(idx); err != nil {
		return SpecIndex{}, err
	}
	return idx, nil
}

func looksLikeURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || (len(s) > 8 && s[:8] == "https://"))
}
EOF

stage_commit "feat(modules): add openapi import/cache skeleton"

# ---------------------------
# 6) Bench module (basic load runner)
# ---------------------------
say "==> Phase 6: bench module v1 skeleton"

cat > internal/modules/bench/module.go <<'EOF'
package bench

import "github.com/bspippi1337/restless/internal/core/app"

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Name() string { return "bench" }

func (m *Module) Register(r *app.Registry) error { return nil }
EOF

cat > internal/modules/bench/bench.go <<'EOF'
package bench

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/bspippi1337/restless/internal/core/types"
)

type Result struct {
	TotalRequests int64
	Errors        int64
	DurationMs    int64
	P50Ms         int64
	P95Ms         int64
	P99Ms         int64
}

type Config struct {
	Concurrency int
	Duration    time.Duration
	Request     types.Request
}

func Run(ctx context.Context, r engine.Runner, cfg Config) (Result, error) {
	if r == nil {
		return Result{}, errors.New("nil runner")
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}
	if cfg.Duration <= 0 {
		cfg.Duration = 5 * time.Second
	}

	deadline := time.Now().Add(cfg.Duration)

	var total int64
	var errs int64
	var mu sync.Mutex
	var durs []int64

	wg := sync.WaitGroup{}
	wg.Add(cfg.Concurrency)

	for i := 0; i < cfg.Concurrency; i++ {
		go func() {
			defer wg.Done()
			for time.Now().Before(deadline) {
				start := time.Now()
				_, err := r.Run(ctx, cfg.Request)
				ms := time.Since(start).Milliseconds()
				atomic.AddInt64(&total, 1)
				if err != nil {
					atomic.AddInt64(&errs, 1)
					continue
				}
				mu.Lock()
				durs = append(durs, ms)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	out := Result{
		TotalRequests: total,
		Errors:        errs,
		DurationMs:    cfg.Duration.Milliseconds(),
	}
	out.P50Ms, out.P95Ms, out.P99Ms = percentiles(durs)
	return out, nil
}

func percentiles(ms []int64) (p50, p95, p99 int64) {
	if len(ms) == 0 {
		return 0, 0, 0
	}
	// simple in-place sort (small enough), avoid extra deps
	for i := 0; i < len(ms); i++ {
		for j := i + 1; j < len(ms); j++ {
			if ms[j] < ms[i] {
				ms[i], ms[j] = ms[j], ms[i]
			}
		}
	}
	get := func(q float64) int64 {
		if len(ms) == 0 {
			return 0
		}
		idx := int(float64(len(ms)-1) * q)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(ms) {
			idx = len(ms) - 1
		}
		return ms[idx]
	}
	return get(0.50), get(0.95), get(0.99)
}
EOF

stage_commit "feat(modules): add bench module v1 skeleton (concurrency runner)"

# ---------------------------
# 7) Export module (report skeleton)
# ---------------------------
say "==> Phase 7: export module skeleton"

cat > internal/modules/export/module.go <<'EOF'
package export

import "github.com/bspippi1337/restless/internal/core/app"

type Module struct{}

func New() *Module { return &Module{} }
func (m *Module) Name() string { return "export" }
func (m *Module) Register(r *app.Registry) error { return nil }
EOF

cat > internal/modules/export/report.go <<'EOF'
package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/bspippi1337/restless/internal/core/types"
)

func SaveJSONArtifact(name string, resp types.Response) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	dir := filepath.Join(home, ".restless", "artifacts", ts)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if name == "" {
		name = "response"
	}
	p := filepath.Join(dir, name+".json")
	b, err := json.MarshalIndent(map[string]any{
		"status":   resp.StatusCode,
		"headers":  resp.Headers,
		"duration": resp.DurationMs,
		"body":     string(resp.Body),
	}, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		return "", err
	}
	return p, nil
}
EOF

stage_commit "feat(modules): add export module skeleton (json artifact)"

# ---------------------------
# 8) CLI v2 command: minimal plumbing using core/app + modules
# ---------------------------
say "==> Phase 8: v2 CLI minimal command wiring"

# We'll add a new cmd: cmd/restless-v2 (non-breaking)
mkdir -p cmd/restless-v2

cat > cmd/restless-v2/main.go <<'EOF'
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/core/types"
	"github.com/bspippi1337/restless/internal/modules/bench"
	"github.com/bspippi1337/restless/internal/modules/export"
	"github.com/bspippi1337/restless/internal/modules/openapi"
	"github.com/bspippi1337/restless/internal/modules/session"
)

func main() {
	var (
		method = flag.String("X", "GET", "HTTP method")
		url    = flag.String("url", "", "Request URL")
		body   = flag.String("d", "", "Body string")
		hdrK   = flag.String("Hk", "", "Header key (single)")
		hdrV   = flag.String("Hv", "", "Header value (single)")
		setVar = flag.String("set", "", "Set session var: key=value")
		doBench = flag.Bool("bench", false, "Run bench mode")
		c       = flag.Int("c", 10, "Bench concurrency")
		dur     = flag.Duration("dur", 5*time.Second, "Bench duration")
		save    = flag.String("save", "", "Save json artifact name")
	)
	flag.Parse()

	// Modules
	sess := session.New()
	_ = openapi.New() // not used yet here, but wired
	_ = export.New()
	_ = bench.New()

	mods := []app.Module{
		sess,
		openapi.New(),
		export.New(),
		bench.New(),
	}

	a, err := app.New(mods)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if *setVar != "" {
		k, v, ok := splitKV(*setVar)
		if !ok {
			fmt.Fprintln(os.Stderr, "invalid -set, want key=value")
			os.Exit(2)
		}
		sess.Set(k, v)
	}

	if *url == "" {
		fmt.Fprintln(os.Stderr, "missing -url")
		flag.Usage()
		os.Exit(2)
	}

	h := http.Header{}
	if *hdrK != "" {
		h.Add(*hdrK, *hdrV)
	}

	req := types.Request{
		Method:  *method,
		URL:     *url,
		Headers: h,
		Body:    []byte(*body),
	}

	ctx := context.Background()

	if *doBench {
		r, err := bench.Run(ctx, aRunner(a), bench.Config{
			Concurrency: *c,
			Duration:    *dur,
			Request:     req,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "bench error:", err)
			os.Exit(1)
		}
		fmt.Printf("bench: total=%d errors=%d dur_ms=%d p50=%d p95=%d p99=%d\n",
			r.TotalRequests, r.Errors, r.DurationMs, r.P50Ms, r.P95Ms, r.P99Ms)
		return
	}

	resp, err := a.RunOnce(ctx, req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request error:", err)
		os.Exit(1)
	}

	fmt.Printf("status: %d (dur=%dms)\n", resp.StatusCode, resp.DurationMs)
	fmt.Printf("%s\n", string(resp.Body))

	if *save != "" {
		p, err := export.SaveJSONArtifact(*save, resp)
		if err != nil {
			fmt.Fprintln(os.Stderr, "save error:", err)
			os.Exit(1)
		}
		fmt.Printf("saved: %s\n", p)
	}
}

func splitKV(s string) (k, v string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
}

// aRunner adapts App.RunOnce into the bench runner interface without exposing internals.
type runOnceAdapter struct{ a *app.App }

func aRunner(a *app.App) *runOnceAdapter { return &runOnceAdapter{a: a} }

func (r *runOnceAdapter) Run(ctx context.Context, req types.Request) (types.Response, error) {
	return r.a.RunOnce(ctx, req)
}
EOF

stage_commit "feat(cli): add cmd/restless-v2 minimal (engine+modules wiring)"

# ---------------------------
# 9) Roadmap refresh
# ---------------------------
say "==> Phase 9: roadmap/docs refresh"

cat > roadmap.md <<'EOF'
# Restless v2 Roadmap

## Strategy
Build a stable v2 core with compile-time modules (single binary), then expand features.
Avoid runtime plugins until the core is boring and stable.

## Principles
- Core is sacred: small, stable, testable.
- Modules evolve fast but must respect boundaries.
- CLI first, TUI premium, GUI last.
- Every phase has exit criteria.

---

## Phase 0: Baseline
- gofmt, tidy, go test ./...
- remove stale entrypoints or isolate behind build tags
- docs: architecture + roadmap aligned with code

Exit:
- `go test ./...` is green

## Phase 1: Engine-first core
- Stable request/response types
- Runner interface
- HTTP runner default
- App wiring surface + hooks (request/response mutators)

Exit:
- CLI can run requests using core app

## Phase 2: Sessions v1
- {{vars}} templating in url/body/headers
- simple JSON dot-path extractor
- regex extractor

Exit:
- example flow demonstrates "extract then reuse"

## Phase 3: OpenAPI v1
- import + cache specs
- list cached specs
- (next) parse endpoints + quick-run

Exit:
- deterministic import/list behavior and docs

## Phase 4: Bench v1
- concurrency run + p50/p95/p99
- output table + json export

Exit:
- bench works reliably, warnings for throttling

## Phase 5: Export/Artifacts
- json artifact saving
- md/html report later

Exit:
- one command makes shareable artifact

## Phase 6: TUI 2.0
- tabs, history, request builder, json viewer

Exit:
- feels like a product (lazygit/k9s vibe)

## Phase 7: GUI
- minimal shell using same core

Exit:
- GUI shares core 100%

## Future: Plugin runtime (WASM)
- define stable plugin interfaces first
- loader later
EOF

stage_commit "docs: refresh roadmap for v2 modules + core"

# ---------------------------
# 10) Examples
# ---------------------------
say "==> Phase 10: examples"

cat > examples/v2_quickstart.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

# Quickstart (v2)
# Build:
#   go build -o restless-v2 ./cmd/restless-v2
# Run:
#   ./restless-v2 -url https://httpbin.org/get
# Template vars:
#   ./restless-v2 -set token=abc123 -Hk Authorization -Hv 'Bearer {{token}}' -url https://httpbin.org/headers
# Bench:
#   ./restless-v2 -bench -c 20 -dur 3s -url https://httpbin.org/get

echo "OK - see comments in this file."
EOF
chmod +x examples/v2_quickstart.sh

stage_commit "docs: add v2 quickstart example script"

# ---------------------------
# 11) Final fmt/tidy/test
# ---------------------------
say "==> Final: fmt/tidy/test"
gofmt -w . >/dev/null 2>&1 || true
go mod tidy >/dev/null 2>&1 || true
go test ./...

say ""
say "✅ Over-the-top v2 scaffolding applied on branch: ${WORK_BRANCH}"
say ""
say "Next commands:"
say "  go build -o restless-v2 ./cmd/restless-v2"
say "  ./restless-v2 -url https://httpbin.org/get"
say "  ./restless-v2 -set token=abc -Hk Authorization -Hv 'Bearer {{token}}' -url https://httpbin.org/headers"
say "  ./restless-v2 -bench -c 20 -dur 3s -url https://httpbin.org/get"
say ""
say "If you want to merge into main:"
say "  git switch main"
say "  git merge ${WORK_BRANCH}"
