package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/bspippi1337/restless/internal/engine"
)

func maybeAutopilot(args []string) bool {

	if len(args) == 0 {
		return false
	}

	target := args[0]

	if !strings.Contains(target, ".") {
		return false
	}

	fmt.Println("Restless autopilot scanning:", target)

	res, err := engine.Run(target)
	if err != nil {
		log.Fatal(err)
	}

	dot := engine.TopologyToDOT(res.Topology)

	out := target + ".svg"

	if err := engine.RenderDOT(dot, out); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Graph written to", out)

	return true
}
