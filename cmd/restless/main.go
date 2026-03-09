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

	cmd.Start()
}

func main() {

	open := false
	target := ""

	for _, a := range os.Args[1:] {

		if a == "--open" || a == "-o" {
			open = true
			continue
		}

		if looksLikeTarget(a) {
			target = a
		}
	}

	if target != "" {

		target = engine.NormalizeTarget(target)

		fmt.Println("Restless API Discovery Engine")
		fmt.Println("Scanning:", target)
		fmt.Println()

		engine.Step(1, 5, "probing API surface")

		res, err := engine.Run(target)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		engine.Step(2, 5, "inferring resource model")

		engine.Step(3, 5, "building topology")

		engine.PrintResult(res)

		engine.Step(4, 5, "generating graph")

		dot := engine.TopologyToDOT(res.Topology)

		out := strings.ReplaceAll(target, "https://", "") + ".svg"

		err = engine.RenderDOT(dot, out)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		engine.Step(5, 5, "complete")

		fmt.Println()
		fmt.Println("Graph written to", out)

		if open {
			openFile(out)
		}

		os.Exit(0)
	}

	cli.Execute()
}
