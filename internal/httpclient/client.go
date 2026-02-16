package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Request struct {
	Method  string
	BaseURL string
	Path    string
	Headers map[string]string
	Query   map[string]string
	Body    []byte
}

type Result struct {
	Status     string
	StatusCode int
	Headers    http.Header
	Body       []byte
	LatencyMs  int64
}

func BuildURL(base, path string, q map[string]string) (string, error) {
	if base == "" {
		return "", errors.New("empty baseURL")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if path != "" {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		u.Path = strings.TrimRight(u.Path, "/") + path
	}
	qs := u.Query()
	for k, v := range q {
		if k != "" {
			qs.Set(k, v)
		}
	}
	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func Do(ctx context.Context, r Request) (Result, error) {
	start := time.Now()
	full, err := BuildURL(r.BaseURL, r.Path, r.Query)
	if err != nil {
		return Result{}, err
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(r.Method), full, bytes.NewReader(r.Body))
	if err != nil {
		return Result{}, err
	}
	for k, v := range r.Headers {
		if k != "" {
			req.Header.Set(k, v)
		}
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "restless/alpha")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	return Result{Status: resp.Status, StatusCode: resp.StatusCode, Headers: resp.Header, Body: b, LatencyMs: time.Since(start).Milliseconds()}, nil
}

func Redact(s string) string {
	pats := []string{os.Getenv("RESTLESS_TOKEN"), os.Getenv("OPENAI_API_KEY")}
	for _, p := range pats {
		if p != "" && len(p) > 8 {
			s = strings.ReplaceAll(s, p, p[:3]+"…"+p[len(p)-3:])
		}
	}
	return s
}

func PrettyJSON(b []byte) []byte {
	var v any
	if json.Unmarshal(b, &v) != nil {
		return b
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return b
	}
	return out
}
