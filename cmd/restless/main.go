package main

import (
	"os"

	"github.com/bspippi1337/restless/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
