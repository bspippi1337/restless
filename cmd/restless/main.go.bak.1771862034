package main

import (
	"os"

	"github.com/bspippi1337/restless/internal/app"
	"github.com/bspippi1337/restless/internal/interactive"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		os.Exit(interactive.Run())
	}

	os.Exit(app.Run(args, os.Stdin, os.Stdout, os.Stderr))
}
