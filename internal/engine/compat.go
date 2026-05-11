package engine

import (
	"strings"
)

func TopologyToDOT(topology string) string {
	lines := strings.Split(topology, "\n")

	var b strings.Builder
	b.WriteString("digraph restless {\n")
	b.WriteString("  rankdir=LR;\n")

	rootExists := false

	for _, raw := range lines {
		line := strings.TrimSpace(raw)

		if line == "" {
			continue
		}

		if line == "root" {
			rootExists = true
			continue
		}

		if !strings.Contains(line, "└──") {
			continue
		}

		parts := strings.SplitN(line, "└──", 2)
		if len(parts) != 2 {
			continue
		}

		node := strings.TrimSpace(parts[1])
		if node == "" {
			continue
		}

		if !rootExists {
			rootExists = true
		}

		safe := strings.ReplaceAll(node, "\"", "")

		b.WriteString("  \"root\" -> \"")
		b.WriteString(safe)
		b.WriteString("\";\n")
	}

	if !rootExists {
		b.WriteString("  \"root\";\n")
	}

	b.WriteString("}\n")

	return b.String()
}
