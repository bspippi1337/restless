package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bspippi1337/restless/internal/cli"
	"github.com/bspippi1337/restless/internal/engine"
)

func looksLikeTarget(s string) bool {
	return strings.Contains(s, ".")
}

func openFile(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	_ = cmd.Start()
}

func main() {

	open := false
	target := ""

	for _, a := range os.Args[1:] {
		if a == "--open" || a == "-o" {
			open = true
		} else if looksLikeTarget(a) {
			target = a
		}
	}

	if target != "" {

		fmt.Println("Restless autopilot scanning:", target)
		fmt.Println()

		res, err := engine.Run(target)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Print CLI report
		engine.PrintResult(res)

		dot := engine.TopologyToDOT(res.Topology)

		out := target + ".svg"

		if err := engine.RenderDOT(dot, out); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println("Graph written to", out)

		if open {
			openFile(out)
		}

		os.Exit(0)
	}

	cli.Execute()
}
