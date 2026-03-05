package recon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Engine struct {
	Client  HTTPClient
	Timeout time.Duration
	UA      string
	Headers map[string]string
}

func New() *Engine {
	return &Engine{
		Client:  &http.Client{Timeout: 6 * time.Second},
		Timeout: 6 * time.Second,
		UA:      "restless-blckswan/1 (+https://github.com/bspippi1337/restless)",
		Headers: map[string]string{},
	}
}

type Response struct {
	URL         string
	Method      string
	Status      int
	Bytes       int
	DurationMS  int64
	ContentType string
	Body        []byte
	Headers     map[string]string
}

func (e *Engine) Request(ctx context.Context, method, rawURL string, body []byte) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", e.UA)
	for k, v := range e.Headers {
		req.Header.Set(k, v)
	}
	start := time.Now()
	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	ct := resp.Header.Get("Content-Type")

	h := map[string]string{}
	for _, k := range []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After"} {
		if v := resp.Header.Get(k); v != "" {
			h[k] = v
		}
	}

	return &Response{
		URL:         rawURL,
		Method:      method,
		Status:      resp.StatusCode,
		Bytes:       len(b),
		DurationMS:  time.Since(start).Milliseconds(),
		ContentType: ct,
		Body:        b,
		Headers:     h,
	}, nil
}

func NormalizeTarget(raw string) (string, *url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil, errors.New("empty target")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" {
		return "", nil, errors.New("target missing host")
	}
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String(), u, nil
}

func LooksJSON(ct string, body []byte) bool {
	ct = strings.ToLower(ct)
	if strings.Contains(ct, "application/json") {
		return true
	}
	if len(body) == 0 {
		return false
	}
	return body[0] == '{' || body[0] == '['
}

func ExtractSameHostPaths(targetHost string, body []byte) []string {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return nil
	}
	out := []string{}
	var walk func(x any)
	walk = func(x any) {
		switch t := x.(type) {
		case map[string]any:
			for _, vv := range t {
				walk(vv)
			}
		case []any:
			for _, vv := range t {
				walk(vv)
			}
		case string:
			s := strings.TrimSpace(t)
			if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
				u, err := url.Parse(s)
				if err == nil && u.Host == targetHost {
					if u.Path != "" {
						out = append(out, u.Path)
					}
				}
			}
		}
	}
	walk(v)

	uniq := map[string]bool{}
	res := []string{}
	for _, p := range out {
		p = "/" + strings.TrimLeft(p, "/")
		if !uniq[p] {
			uniq[p] = true
			res = append(res, p)
		}
	}
	return res
}
