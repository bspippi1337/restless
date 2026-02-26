package app

import (
	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/bspippi1337/restless/internal/core/logx"
)

// Module is a compile-time feature module.
// Future plugin runtimes can implement the same interface (WASM, etc).
type Module interface {
	Name() string
	Register(r *Registry) error
}

// Registry is the wiring surface between core and modules.
// Keep it stable and small.
type Registry struct {
	Log    *logx.Logger
	Runner engine.Runner

	// Hooks
	RequestMutators  []func(*RequestContext) error
	ResponseMutators []func(*ResponseContext) error
}

type RequestContext struct {
	// For future: env/profile/session vars etc
	Method string
	URL    string
	Body   []byte
	Header map[string][]string
}

type ResponseContext struct {
	StatusCode int
	Body       []byte
	Header     map[string][]string
}

func NewRegistry(lg *logx.Logger, runner engine.Runner) *Registry {
	if lg == nil {
		lg = logx.New(logx.Info)
	}
	return &Registry{
		Log:    lg,
		Runner: runner,
	}
}
