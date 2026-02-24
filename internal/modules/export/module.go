package export

import "github.com/bspippi1337/restless/internal/core/app"

type Module struct{}

func New() *Module { return &Module{} }
func (m *Module) Name() string { return "export" }
func (m *Module) Register(r *app.Registry) error { return nil }
