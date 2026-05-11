package magiswarm

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

type Options struct {
	TargetURL     string
	Concurrency   int
	MaxRequests   int
	Timeout       time.Duration
	Wordlist      []string
	Headers       map[string]string
	EnableFuzz    bool
	FuzzParams    []string
	FuzzValues    []string
	UserAgent     string
	RespectHost   bool
	IncludeRoot   bool
	IncludeCommon bool
}

type Endpoint struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Status      int               `json:"status"`
	ContentType string            `json:"content_type,omitempty"`
	Bytes       int               `json:"bytes,omitempty"`
	DurationMS  int64             `json:"duration_ms,omitempty"`
	Notes       []string          `json:"notes,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type Report struct {
	Target      string     `json:"target"`
	GeneratedAt string     `json:"generated_at"`
	Options     OptionsOut `json:"options"`
	Endpoints   []Endpoint `json:"endpoints"`
	Stats       Stats      `json:"stats"`
	Topology    string     `json:"topology_ascii"`
	Warnings    []string   `json:"warnings,omitempty"`
}

type OptionsOut struct {
	Concurrency int    `json:"concurrency"`
	MaxRequests int    `json:"max_requests"`
	TimeoutMS   int64  `json:"timeout_ms"`
	EnableFuzz  bool   `json:"enable_fuzz"`
	WordlistN   int    `json:"wordlist_n"`
	UserAgent   string `json:"user_agent"`
}

type Stats struct {
	Requests int `json:"requests"`
	Found    int `json:"found"`
	Errors   int `json:"errors"`
	Unique   int `json:"unique_paths"`
}

type Runner struct {
	opt Options
	hc  *http.Client
	u   *url.URL

	mu       sync.Mutex
	seen     map[string]bool
	found    map[string]Endpoint
	queue    []string
	requests int
	errors   int
	warnings []string
}

func DefaultOptions(target string) Options {
	return Options{
		TargetURL:     target,
		Concurrency:   8,
		MaxRequests:   200,
		Timeout:       5 * time.Second,
		Wordlist:      nil,
		Headers:       map[string]string{},
		EnableFuzz:    true,
		FuzzParams:    []string{"id", "q", "search", "debug", "test"},
		FuzzValues:    []string{"1", "0", "true", "test", "admin"},
		UserAgent:     "restless-magiswarm/1 (+https://github.com/bspippi1337/restless)",
		RespectHost:   true,
		IncludeRoot:   true,
		IncludeCommon: true,
	}
}

func New(opt Options) (*Runner, error) {
	if opt.Concurrency <= 0 {
		opt.Concurrency = 8
	}
	if opt.MaxRequests <= 0 {
		opt.MaxRequests = 200
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 5 * time.Second
	}
	if opt.UserAgent == "" {
		opt.UserAgent = "restless-magiswarm/1"
	}

	u, err := url.Parse(opt.TargetURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid target url: missing host")
	}

	hc := &http.Client{Timeout: opt.Timeout}

	r := &Runner{
		opt:   opt,
		hc:    hc,
		u:     u,
		seen:  map[string]bool{},
		found: map[string]Endpoint{},
		queue: []string{},
	}
	return r, nil
}

func (r *Runner) enqueue(p string) {
	p = "/" + strings.TrimLeft(p, "/")
	p = path.Clean(p)
	if p == "." {
		p = "/"
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.seen[p] {
		return
	}
	r.seen[p] = true
	r.queue = append(r.queue, p)
}

func (r *Runner) pop() (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.queue) == 0 {
		return "", false
	}
	p := r.queue[0]
	r.queue = r.queue[1:]
	return p, true
}

func (r *Runner) addFound(ep Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := ep.Method + " " + ep.Path
	r.found[key] = ep
}

func (r *Runner) incReq() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests++
	return r.requests <= r.opt.MaxRequests
}

func (r *Runner) addErr() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors++
}

func (r *Runner) warn(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warnings = append(r.warnings, msg)
}

func (r *Runner) makeURL(p string) string {
	u := *r.u
	u.Path = path.Join(r.u.Path, p)
	if p == "/" {
		u.Path = r.u.Path
		if u.Path == "" {
			u.Path = "/"
		}
	}
	return u.String()
}

func (r *Runner) request(ctx context.Context, method, p string, query url.Values) (Endpoint, []byte, error) {
	start := time.Now()
	full := r.makeURL(p)

	uu, _ := url.Parse(full)
	if query != nil {
		uu.RawQuery = query.Encode()
		full = uu.String()
	}

	req, _ := http.NewRequestWithContext(ctx, method, full, nil)
	req.Header.Set("User-Agent", r.opt.UserAgent)
	for k, v := range r.opt.Headers {
		req.Header.Set(k, v)
	}

	resp, err := r.hc.Do(req)
	if err != nil {
		return Endpoint{Path: p, Method: method}, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))

	ep := Endpoint{
		Path:        p,
		Method:      method,
		Status:      resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Bytes:       len(body),
		DurationMS:  time.Since(start).Milliseconds(),
		Headers:     map[string]string{},
	}

	for _, hk := range []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After"} {
		if hv := resp.Header.Get(hk); hv != "" {
			ep.Headers[hk] = hv
		}
	}
	if len(ep.Headers) > 0 {
		ep.Notes = append(ep.Notes, "rate-limit-headers")
	}
	return ep, body, nil
}

func (r *Runner) extractPathsFromJSON(body []byte) []string {
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil
	}

	out := []string{}
	host := r.u.Host

	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			for _, vv := range t {
				walk(vv)
			}
		case []any:
			for _, vv := range t {
				walk(vv)
			}
		case string:
			s := t
			if strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://") {
				uu, err := url.Parse(s)
				if err == nil && uu.Host == host {
					out = append(out, uu.Path)
				}
			}
		}
	}
	walk(obj)

	uniq := map[string]bool{}
	clean := []string{}
	for _, p := range out {
		p = path.Clean("/" + strings.TrimLeft(p, "/"))
		if p == "." {
			p = "/"
		}
		if p == "" {
			continue
		}
		if !uniq[p] {
			uniq[p] = true
			clean = append(clean, p)
		}
	}
	sort.Strings(clean)
	return clean
}

func (r *Runner) seeds() []string {
	seeds := []string{}
	if r.opt.IncludeRoot {
		seeds = append(seeds, "/")
	}
	if r.opt.IncludeCommon {
		seeds = append(seeds,
			"/api", "/v1", "/v2", "/v3", "/graphql",
			"/openapi.json", "/swagger.json", "/swagger/v1/swagger.json",
			"/health", "/status", "/version",
			"/users", "/repos", "/orgs", "/search",
		)
	}
	for _, w := range r.opt.Wordlist {
		w = strings.TrimSpace(w)
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		if !strings.HasPrefix(w, "/") {
			w = "/" + w
		}
		seeds = append(seeds, w)
	}
	return seeds
}

func (r *Runner) Run(ctx context.Context) (*Report, error) {
	for _, p := range r.seeds() {
		r.enqueue(p)
	}

	workers := r.opt.Concurrency
	var wg sync.WaitGroup

	type job struct{ path string }
	jobs := make(chan job)

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			if !r.incReq() {
				return
			}

			ep, body, err := r.request(ctx, "GET", j.path, nil)
			if err != nil {
				r.addErr()
				continue
			}

			if ep.Status < 500 {
				r.addFound(ep)
				ct := strings.ToLower(ep.ContentType)
				if strings.Contains(ct, "application/json") && len(body) > 0 {
					for _, p := range r.extractPathsFromJSON(body) {
						r.enqueue(p)
					}
				}
			}

			if r.opt.EnableFuzz && ep.Status < 500 {
				r.fuzzOne(ctx, j.path)
			}
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker()
	}

	for {
		p, ok := r.pop()
		if !ok {
			break
		}
		jobs <- job{path: p}
	}
	close(jobs)
	wg.Wait()

	found := r.snapshotFound()
	top := BuildTopologyASCII(r.u.Host, found)

	rep := &Report{
		Target:      r.u.String(),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Options: OptionsOut{
			Concurrency: r.opt.Concurrency,
			MaxRequests: r.opt.MaxRequests,
			TimeoutMS:   r.opt.Timeout.Milliseconds(),
			EnableFuzz:  r.opt.EnableFuzz,
			WordlistN:   len(r.opt.Wordlist),
			UserAgent:   r.opt.UserAgent,
		},
		Endpoints: found,
		Stats: Stats{
			Requests: r.requests,
			Found:    len(found),
			Errors:   r.errors,
			Unique:   countUniquePaths(found),
		},
		Topology: top,
		Warnings: r.warnings,
	}

	return rep, nil
}

func (r *Runner) snapshotFound() []Endpoint {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Endpoint, 0, len(r.found))
	for _, ep := range r.found {
		out = append(out, ep)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Method < out[j].Method
		}
		return out[i].Path < out[j].Path
	})
	return out
}

func countUniquePaths(eps []Endpoint) int {
	uniq := map[string]bool{}
	for _, e := range eps {
		uniq[e.Path] = true
	}
	return len(uniq)
}

func (r *Runner) fuzzOne(ctx context.Context, p string) {
	if len(r.opt.FuzzParams) == 0 || len(r.opt.FuzzValues) == 0 {
		return
	}
	seed := sha1.Sum([]byte(p))
	ixp := int(seed[0]) % len(r.opt.FuzzParams)
	ixv := int(seed[1]) % len(r.opt.FuzzValues)

	q := url.Values{}
	q.Set(r.opt.FuzzParams[ixp], r.opt.FuzzValues[ixv])

	ep, _, err := r.request(ctx, "GET", p, q)
	if err != nil {
		return
	}
	if ep.Status < 500 && ep.Status != 404 {
		ep.Notes = append(ep.Notes, "fuzz-hit")
		r.addFound(ep)
	}
}

func WriteReportFiles(rep *Report, outDir string) (jsonPath, topologyPath string, err error) {
	if outDir == "" {
		outDir = "dist"
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", "", err
	}

	host := "target"
	if uu, e := url.Parse(rep.Target); e == nil && uu.Host != "" {
		host = uu.Host
	}
	stamp := time.Now().UTC().Format("20060102_150405")
	base := fmt.Sprintf("magiswarm_%s_%s", host, stamp)

	jsonPath = path.Join(outDir, base+".json")
	topologyPath = path.Join(outDir, base+".topology.txt")

	b, _ := json.MarshalIndent(rep, "", "  ")
	if err := os.WriteFile(jsonPath, b, 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(topologyPath, []byte(rep.Topology+"\n"), 0o644); err != nil {
		return "", "", err
	}
	return jsonPath, topologyPath, nil
}

func ShortID(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])[:8]
}
