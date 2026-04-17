package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bspippi1337/restless/internal/cli"
	"github.com/bspippi1337/restless/internal/engine"
)

func looksLikeTarget(s string) bool {
	if strings.Contains(s, ".") {
		return true
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}
	return false
}

func runTarget(target string) {
	engine.Step(1, 4, "probing API surface")

	norm := engine.NormalizeTarget(target)

	engine.Step(2, 4, "inferring structure")

	res, err := engine.Run(norm)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	engine.Step(3, 4, "building topology")

	engine.Print(res)

	engine.Step(4, 4, "rendering graph")

	out := strings.ReplaceAll(target, "https://", "")
	out = strings.ReplaceAll(out, "http://", "")
	out += ".svg"

	dot := engine.TopologyToDOT(res.Topology)
	_ = engine.RenderDOT(dot, out)

	fmt.Println()
	fmt.Println("Graph written:", out)
}

func main() {
	if len(os.Args) == 2 && looksLikeTarget(os.Args[1]) {
		runTarget(os.Args[1])
		return
	}

	cli.Execute()
}
