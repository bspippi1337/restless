package httpx

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	StatusCode int
	Body       []byte
	Latency    time.Duration
	Headers    http.Header

	RateLimitRemaining int
	RateLimitReset     int64
}

type Executor struct {
	client *http.Client
}

func NewExecutor(timeout time.Duration) *Executor {
	return &Executor{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (e *Executor) Do(method, url string, body []byte) (*Response, error) {

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	r := &Response{
		StatusCode: resp.StatusCode,
		Body:       data,
		Latency:    time.Since(start),
		Headers:    resp.Header,
	}

	if v := resp.Header.Get("X-RateLimit-Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			r.RateLimitRemaining = n
		}
	}

	if v := resp.Header.Get("X-RateLimit-Reset"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			r.RateLimitReset = n
		}
	}

	return r, nil
}
