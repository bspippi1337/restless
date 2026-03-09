package engine

import (
	"fmt"
	"strings"
)

func TopologyToDOT(topology string) string {

	lines := strings.Split(topology, "\n")

	var edges []string
	var stack []string

	for _, line := range lines {

		level := strings.Count(line, "  ")
		node := strings.TrimSpace(strings.ReplaceAll(line, "└──", ""))

		if node == "" || node == "root" {
			stack = []string{"root"}
			continue
		}

		if level+1 <= len(stack) {
			stack = stack[:level+1]
		}

		parent := stack[len(stack)-1]
		edges = append(edges, fmt.Sprintf(`"%s" -> "%s"`, parent, node))

		stack = append(stack, node)
	}

	out := "digraph API {\n"
	out += "  rankdir=LR;\n"
	out += "  node [shape=box, style=rounded];\n"

	for _, e := range edges {
		out += "  " + e + ";\n"
	}

	out += "}\n"

	return out
}
