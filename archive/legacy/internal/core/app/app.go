package app

import (
	"context"
	"net/http"

	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/bspippi1337/restless/internal/core/logx"
	"github.com/bspippi1337/restless/internal/core/types"
)

type App struct {
	reg *Registry
}

func New(mods []Module) (*App, error) {
	lg := logx.New(logx.Info)
	runner := engine.NewHTTPRunner(nil)
	reg := NewRegistry(lg, runner)

	for _, m := range mods {
		reg.Log.Printf(logx.Info, "registering module: %s", m.Name())
		if err := m.Register(reg); err != nil {
			return nil, err
		}
	}

	return &App{reg: reg}, nil
}

func (a *App) RunOnce(ctx context.Context, req types.Request) (types.Response, error) {
	// Apply request mutators (future: template vars, auth, etc)
	rc := &RequestContext{
		Method: req.Method,
		URL:    req.URL,
		Body:   req.Body,
		Header: map[string][]string{},
	}
	if req.Headers != nil {
		for k, vv := range req.Headers {
			rc.Header[k] = append([]string{}, vv...)
		}
	}
	for _, fn := range a.reg.RequestMutators {
		if err := fn(rc); err != nil {
			return types.Response{}, err
		}
	}

	// Build final request
	h := http.Header{}
	for k, vv := range rc.Header {
		for _, v := range vv {
			h.Add(k, v)
		}
	}
	finalReq := types.Request{
		Method:  rc.Method,
		URL:     rc.URL,
		Headers: h,
		Body:    rc.Body,
	}

	resp, err := a.reg.Runner.Run(ctx, finalReq)
	if err != nil {
		return types.Response{}, err
	}

	// Apply response mutators
	rsc := &ResponseContext{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Header:     map[string][]string{},
	}
	for k, vv := range resp.Headers {
		rsc.Header[k] = append([]string{}, vv...)
	}
	for _, fn := range a.reg.ResponseMutators {
		if err := fn(rsc); err != nil {
			return types.Response{}, err
		}
	}

	// Write back
	outHeaders := http.Header{}
	for k, vv := range rsc.Header {
		for _, v := range vv {
			outHeaders.Add(k, v)
		}
	}
	return types.Response{
		StatusCode: rsc.StatusCode,
		Headers:    outHeaders,
		Body:       rsc.Body,
		DurationMs: resp.DurationMs,
	}, nil
}
