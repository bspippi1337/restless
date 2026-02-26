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
