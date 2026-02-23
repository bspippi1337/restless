package app

import (
	"context"
	"fmt"
)

type Mode string

const (
	ModeNormal Mode = "normal"
	ModeSmart  Mode = "smart"
)

type Options struct {
	Mode Mode
	Args []string
}

// TODO: Wire this into your existing core (client/tui/openapi/etc).
func Run(ctx context.Context, opts Options) error {
	switch opts.Mode {
	case ModeSmart:
		// TODO: call your previous restless-smart entry logic here
		return fmt.Errorf("smart mode not wired yet (opts.Args=%v)", opts.Args)
	case ModeNormal:
		// TODO: call your previous restless entry logic here
		return fmt.Errorf("normal mode not wired yet (opts.Args=%v)", opts.Args)
	default:
		return fmt.Errorf("unknown mode: %q", opts.Mode)
	}
}
