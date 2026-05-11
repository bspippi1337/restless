package cli

import (
	"fmt"
	"os"

	"github.com/bspippi1337/restless/internal/engine"
)

func maybeAutopilot(args []string) {

	if len(args) == 0 {
		return
	}

	target := args[len(args)-1]

	if target == "" {
		return
	}

	if args[0] != "inspect" &&
		args[0] != "scan" &&
		args[0] != "discover" {
		return
	}

	fmt.Fprintf(os.Stderr, "Restless autopilot scanning: %s\n", target)

	res, err := engine.Run(target)
	if err != nil {
		return
	}

	if res.Topology != "" {
		dot := engine.TopologyToDOT(res.Topology)
		engine.RenderDOT(dot)
	}
}
