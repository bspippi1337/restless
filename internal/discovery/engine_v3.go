package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	Client    *http.Client
	Base      *url.URL
	Visited   map[string]bool
	Endpoints map[string]*Endpoint
	Queue     chan string
	Workers   int
	MaxDepth  int
	mu        sync.Mutex
	wg        sync.WaitGroup
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

func NewEngine(raw string) (*Engine, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	e := &Engine{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Base:      u,
		Visited:   map[string]bool{},
		Endpoints: map[string]*Endpoint{},
		Queue:     make(chan string, 1024),
		Workers:   6,
		MaxDepth:  4,
	}

	return e, nil
}

func (e *Engine) Discover(ctx context.Context) map[string]*Endpoint {
	e.Queue <- e.Base.Path

	for i := 0; i < e.Workers; i++ {
		e.wg.Add(1)
		go e.worker(ctx)
	}

	e.wg.Wait()
	return e.Endpoints
}

func (e *Engine) worker(ctx context.Context) {
	defer e.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case path := <-e.Queue:
			if e.shouldSkip(path) {
				continue
			}

			e.scanEndpoint(ctx, path)
		}
	}
}

func (e *Engine) shouldSkip(path string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Visited[path] {
		return true
	}

	e.Visited[path] = true
	return false
}

func (e *Engine) scanEndpoint(ctx context.Context, path string) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}

	for _, m := range methods {
		resp, body, err := e.doRequest(ctx, m, path)
		if err != nil {
			continue
		}

		e.registerEndpoint(path, m, resp.StatusCode, body)
		e.extractLinks(path, body)
	}
}

func (e *Engine) doRequest(ctx context.Context, method, path string) (*http.Response, []byte, error) {
	u := *e.Base
	u.Path = path

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	return resp, body, nil
}

func (e *Engine) registerEndpoint(path, method string, status int, body []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ep, ok := e.Endpoints[path]
	if !ok {
		ep = &Endpoint{
			Path:       path,
			Methods:    map[string]*MethodInfo{},
			Parameters: map[string]string{},
		}
		e.Endpoints[path] = ep
	}

	schema := detectSchema(body)

	ep.Methods[method] = &MethodInfo{
		Status: status,
		Schema: schema,
	}

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

	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			out[k] = typeOf(val)
		}

	case []interface{}:
		if len(v) > 0 {
			if m, ok := v[0].(map[string]interface{}); ok {
				for k, val := range m {
					out[k] = typeOf(val)
				}
			}
		}
	}

	return out
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
	default:
		return "unknown"
	}
}

func (e *Engine) extractLinks(base string, body []byte) {
	var data interface{}

	if err := json.Unmarshal(body, &data); err != nil {
		return
	}

	links := collectLinks(data)

	for _, l := range links {
		if !strings.HasPrefix(l, "/") {
			continue
		}

		e.Queue <- normalizePath(base, l)
	}
}

func collectLinks(v interface{}) []string {
	out := []string{}

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

var endpointRegex = regexp.MustCompile(`^/[\w/\-]+`)

func looksLikeEndpoint(s string) bool {
	return endpointRegex.MatchString(s)
}

func normalizePath(base, found string) string {
	if strings.HasPrefix(found, base) {
		return found
	}

	if strings.HasSuffix(base, "/") {
		return base + strings.TrimPrefix(found, "/")
	}

	return base + found
}

func (e *Engine) PrintMap() {
	for path, ep := range e.Endpoints {
		fmt.Println(path)

		for m, info := range ep.Methods {
			fmt.Printf("  %s -> %d\n", m, info.Status)
		}

		for p, t := range ep.Parameters {
			fmt.Printf("    %s : %s\n", p, t)
		}
	}
}
