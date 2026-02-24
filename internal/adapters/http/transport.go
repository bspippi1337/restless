package httpadapter

import (
	"bytes"
	"context"
	"github.com/bspippi1337/restless/internal/core/engine"
	"io"
	"net/http"
)

type HTTPTransport struct{}

func (t *HTTPTransport) Do(ctx context.Context, job engine.Job) engine.Result {
	req, err := http.NewRequestWithContext(ctx, job.Method, job.Target, bytes.NewReader(job.Body))
	if err != nil {
		return engine.Result{Err: err}
	}

	for k, v := range job.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return engine.Result{Err: err}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	return engine.Result{
		Status: resp.StatusCode,
		Body:   body,
	}
}
