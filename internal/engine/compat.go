package engine

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Print(r *Result) {
	PrintResult(r)
}

func TopologyToDOT(topology string) string {

	lines := strings.Split(topology, "\n")

	var dot strings.Builder
	dot.WriteString("digraph API {\n")
	dot.WriteString("rankdir=LR;\n")

	var stack []string

	for _, line := range lines {

		level := strings.Count(line, "  ")
		name := strings.TrimSpace(strings.ReplaceAll(line, "└──", ""))

		if name == "" {
			continue
		}

		if name == "root" {
			stack = []string{"root"}
			continue
		}

		if level < len(stack) {
			stack = stack[:level]
		}

		parent := stack[len(stack)-1]

		dot.WriteString(fmt.Sprintf("\"%s\" -> \"%s\";\n", parent, name))

		stack = append(stack, name)

	}

	dot.WriteString("}\n")

	return dot.String()
}

func RenderDOT(dot string, file string) error {

	tmp := file + ".dot"

	os.WriteFile(tmp, []byte(dot), 0644)

	cmd := exec.Command("dot", "-Tsvg", tmp, "-o", file)

	return cmd.Run()

}
