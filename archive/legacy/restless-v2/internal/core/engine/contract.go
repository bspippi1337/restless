package engine

import (
	"context"

	"github.com/bspippi1337/restless/internal/core/types"
)

// Runner executes a request and returns a normalized response.
// This is the stable "spine" of Restless v2.
type Runner interface {
	Run(ctx context.Context, req types.Request) (types.Response, error)
}
