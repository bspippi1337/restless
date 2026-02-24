package app

import (
    "context"
    "github.com/bspippi1337/restless/internal/core/engine"
)

type Runner struct {
    Engine *engine.Engine
}

func (r *Runner) Run(ctx context.Context, method, target string) engine.Result {
    job := engine.Job{
        Method: method,
        Target: target,
        Headers: map[string]string{
            "User-Agent": "restless-v2",
        },
    }
    return r.Engine.Run(ctx, job)
}
