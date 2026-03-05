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
