package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	Client    *http.Client
	Base      *url.URL
	Visited   map[string]bool
	Endpoints map[string]*Endpoint
	Workers   int
	MaxDepth  int
	mu        sync.Mutex
}

type Endpoint struct {
	Path       string
	Methods    map[string]*MethodInfo
	Parameters map[string]string
	Children   []string
}

type MethodInfo struct {
	Status int
	Schema map[string]string
}

type crawlItem struct {
	Path  string
	Depth int
}

func NewEngine(raw string) (*Engine, error) {
	u, err := normalizeBaseURL(raw)
	if err != nil {
		return nil, err
	}

	return &Engine{
		Client: &http.Client{Timeout: 10 * time.Second},
		Base:      u,
		Visited:   map[string]bool{},
		Endpoints: map[string]*Endpoint{},
		Workers:   6,
		MaxDepth:  4,
	}, nil
}

func normalizeBaseURL(raw string) (*url.URL, error) {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid target: %s", raw)
	}
	if u.Path == "" {
		u.Path = "/"
	}
	return u, nil
}

func (e *Engine) Discover(ctx context.Context) map[string]*Endpoint {
	frontier := []crawlItem{{Path: cleanPath(e.Base.Path), Depth: 0}}

	for len(frontier) > 0 {
		select {
		case <-ctx.Done():
			return e.Endpoints
		default:
		}

		item := frontier[0]
		frontier = frontier[1:]

		if item.Depth > e.MaxDepth || e.shouldSkip(item.Path) {
			continue
		}

		children := e.scanEndpoint(ctx, item.Path)
		for _, child := range children {
			child = cleanPath(child)
			if child == "" || child == item.Path {
				continue
			}
			if !e.seen(child) {
				frontier = append(frontier, crawlItem{Path: child, Depth: item.Depth + 1})
			}
		}
	}

	return e.Endpoints
}

func (e *Engine) shouldSkip(path string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	path = cleanPath(path)
	if e.Visited[path] {
		return true
	}
	e.Visited[path] = true
	return false
}

func (e *Engine) seen(path string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Visited[cleanPath(path)]
}

func (e *Engine) scanEndpoint(ctx context.Context, path string) []string {
	methods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	var children []string

	for _, method := range methods {
		resp, body, err := e.doRequest(ctx, method, path)
		if err != nil || resp == nil {
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			continue
		}
		e.registerEndpoint(path, method, resp.StatusCode, body)
		children = append(children, e.extractLinks(path, body)...)
	}

	return uniqueStrings(children)
}

func (e *Engine) doRequest(ctx context.Context, method, path string) (*http.Response, []byte, error) {
	u := *e.Base
	u.Path = cleanPath(path)
	u.RawQuery = ""
	u.Fragment = ""

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "restless-api-discovery/1.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp, body, nil
}

func (e *Engine) registerEndpoint(path, method string, status int, body []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()

	path = cleanPath(path)
	ep, ok := e.Endpoints[path]
	if !ok {
		ep = &Endpoint{Path: path, Methods: map[string]*MethodInfo{}, Parameters: map[string]string{}}
		e.Endpoints[path] = ep
	}

	schema := detectSchema(body)
	ep.Methods[method] = &MethodInfo{Status: status, Schema: schema}
	for k, v := range schema {
		if _, ok := ep.Parameters[k]; !ok {
			ep.Parameters[k] = v
		}
	}
}

func detectSchema(body []byte) map[string]string {
	out := map[string]string{}
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return out
	}
	flattenSchema("", data, out, 2)
	return out
}

func flattenSchema(prefix string, v interface{}, out map[string]string, depth int) {
	if depth < 0 {
		return
	}
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			out[key] = typeOf(val)
			flattenSchema(key, val, out, depth-1)
		}
	case []interface{}:
		out[prefix] = "array"
		if len(t) > 0 {
			flattenSchema(prefix+"[]", t[0], out, depth-1)
		}
	}
}

func typeOf(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "bool"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

func (e *Engine) extractLinks(base string, body []byte) []string {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}
	links := collectLinks(data)
	out := make([]string, 0, len(links))
	for _, l := range links {
		p, ok := e.toLocalPath(base, l)
		if ok {
			out = append(out, p)
		}
	}
	return out
}

func collectLinks(v interface{}) []string {
	var out []string
	switch t := v.(type) {
	case map[string]interface{}:
		for _, val := range t {
			out = append(out, collectLinks(val)...)
		}
	case []interface{}:
		for _, val := range t {
			out = append(out, collectLinks(val)...)
		}
	case string:
		if looksLikeEndpoint(t) {
			out = append(out, t)
		}
	}
	return out
}

var endpointRegex = regexp.MustCompile(`^(https?://|/)[A-Za-z0-9_./?=&%-]+$`)

func looksLikeEndpoint(s string) bool {
	return endpointRegex.MatchString(strings.TrimSpace(s))
}

func (e *Engine) toLocalPath(base, found string) (string, bool) {
	found = strings.TrimSpace(found)
	if found == "" {
		return "", false
	}
	if strings.HasPrefix(found, "http://") || strings.HasPrefix(found, "https://") {
		u, err := url.Parse(found)
		if err != nil || u.Host != e.Base.Host {
			return "", false
		}
		return cleanPath(u.Path), true
	}
	if strings.HasPrefix(found, "/") {
		return cleanPath(found), true
	}
	return cleanPath(strings.TrimRight(base, "/") + "/" + found), true
}

func cleanPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = cleanPath(s)
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func (e *Engine) PrintMap() {
	paths := make([]string, 0, len(e.Endpoints))
	for path := range e.Endpoints {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		ep := e.Endpoints[path]
		fmt.Println(path)
		methods := make([]string, 0, len(ep.Methods))
		for m := range ep.Methods {
			methods = append(methods, m)
		}
		sort.Strings(methods)
		for _, m := range methods {
			fmt.Printf("  %s -> %d\n", m, ep.Methods[m].Status)
		}
		fields := make([]string, 0, len(ep.Parameters))
		for p := range ep.Parameters {
			fields = append(fields, p)
		}
		sort.Strings(fields)
		for _, p := range fields {
			fmt.Printf("    %s : %s\n", p, ep.Parameters[p])
		}
	}
}
