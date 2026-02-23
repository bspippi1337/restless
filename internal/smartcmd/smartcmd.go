package smartcmd

import (
	"fmt"
	"os"

	"github.com/bspippi1337/restless/internal/app"
	"github.com/bspippi1337/restless/internal/discover"
	"github.com/bspippi1337/restless/internal/simulator"
)

func Dispatch(args []string) (bool, int) {
	if len(args) == 0 {
		return false, 0
	}

	switch args[0] {

	case "probe":
		if len(args) < 2 {
			fmt.Println("usage: restless probe <url>")
			return true, 2
		}
		p, err := discover.Probe(args[1])
		if err != nil {
			fmt.Println("probe error:", err)
			return true, 1
		}
		os.Stdout.Write(p.JSON())
		return true, 0

	case "simulate":
		if len(args) < 2 {
			fmt.Println("usage: restless simulate <url>")
			return true, 2
		}
		method, url := simulator.Run(args[1])
		return true, app.Run([]string{method, url}, os.Stdin, os.Stdout, os.Stderr)
	}

	return false, 0
}
