package engine

import "context"

type Job struct {
    Target  string
    Method  string
    Body    []byte
    Headers map[string]string
}

type Result struct {
    Status int
    Body   []byte
    Err    error
}

type Transport interface {
    Do(ctx context.Context, job Job) Result
}

type Engine struct {
    Transport Transport
}

func (e *Engine) Run(ctx context.Context, job Job) Result {
    return e.Transport.Do(ctx, job)
}
