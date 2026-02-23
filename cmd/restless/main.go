package main

import (
	"os"

	"github.com/bspippi1337/restless/internal/app"
	"github.com/bspippi1337/restless/internal/interactive"
	"github.com/bspippi1337/restless/internal/smartcmd"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		os.Exit(interactive.Run())
	}

	if handled, code := smartcmd.Dispatch(args); handled {
		os.Exit(code)
	}

	os.Exit(app.Run(args, os.Stdin, os.Stdout, os.Stderr))
}
