package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bspippi1337/restless/internal/app"
)

func main() {
	args := os.Args[1:]
	mode := app.ModeNormal

	// explicit flags override
	for i := 0; i < len(args); i++ {
		if args[i] == "--smart" {
			mode = app.ModeSmart
			args = append(args[:i], args[i+1:]...)
			i--
		} else if args[i] == "--normal" {
			mode = app.ModeNormal
			args = append(args[:i], args[i+1:]...)
			i--
		}
	}

	// subcommand
	if len(args) > 0 && args[0] == "smart" {
		mode = app.ModeSmart
		args = args[1:]
	}

	if err := app.Run(context.Background(), app.Options{Mode: mode, Args: args}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
