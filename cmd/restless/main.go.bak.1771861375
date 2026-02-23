package main

import (
	"github.com/bspippi1337/restless/internal/app"
	"github.com/bspippi1337/restless/internal/tui"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		if err := tui.Start(os.Stdin, os.Stdout); err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
			os.Exit(1)
		}
		return
	}
	os.Exit(app.Run(args, os.Stdin, os.Stdout, os.Stderr))
}
