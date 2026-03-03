package httpx

import (
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

func (e *Executor) Do(method, url string) (*Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	latency := time.Since(start)

	r := &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Latency:    latency,
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
