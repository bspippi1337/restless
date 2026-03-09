package engine

import (
	"fmt"
	"strings"
)

func TopologyToDOT(topology string) string {

	lines := strings.Split(topology, "\n")

	var edges []string
	var nodes []string
	var stack []string

	for _, line := range lines {

		level := strings.Count(line, "  ")
		node := strings.TrimSpace(strings.ReplaceAll(line, "└──", ""))

		if node == "" {
			continue
		}

		if node == "root" {
			stack = []string{"root"}
			nodes = append(nodes, fmt.Sprintf(`"%s" [%s];`, node, nodeStyle(node)))
			continue
		}

		if level+1 <= len(stack) {
			stack = stack[:level+1]
		}

		parent := stack[len(stack)-1]

		nodes = append(nodes, fmt.Sprintf(`"%s" [%s];`, node, nodeStyle(node)))
		edges = append(edges, fmt.Sprintf(`"%s" -> "%s";`, parent, node))

		stack = append(stack, node)
	}

	out := "digraph API {\n"
	out += "rankdir=LR;\n"
	out += "node [fontname=Helvetica];\n"

	for _, n := range nodes {
		out += n + "\n"
	}

	for _, e := range edges {
		out += e + "\n"
	}

	out += "}\n"

	return out
}
