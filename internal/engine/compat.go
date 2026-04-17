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
	_ = os.WriteFile(tmp, []byte(dot), 0o644)

	cmd := exec.Command("dot", "-Tsvg", tmp, "-o", file)
	if err := cmd.Run(); err == nil {
		return nil
	}

	return renderFallbackSVG(dot, file)
}

func renderFallbackSVG(dot string, file string) error {
	lines := strings.Split(dot, "\n")
	nodes := make([]string, 0, len(lines))
	seen := map[string]bool{}

	for _, line := range lines {
		if !strings.Contains(line, "->") {
			continue
		}
		parts := strings.Split(line, "->")
		if len(parts) != 2 {
			continue
		}
		left := cleanNode(parts[0])
		right := cleanNode(parts[1])
		if left != "" && !seen[left] {
			seen[left] = true
			nodes = append(nodes, left)
		}
		if right != "" && !seen[right] {
			seen[right] = true
			nodes = append(nodes, right)
		}
	}

	if len(nodes) == 0 {
		nodes = []string{"root"}
	}

	height := 80 + len(nodes)*38
	var svg strings.Builder
	svg.WriteString(fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"960\" height=\"%d\" viewBox=\"0 0 960 %d\">", height, height))
	svg.WriteString("<rect width=\"100%\" height=\"100%\" fill=\"#0b0f19\"/>")
	svg.WriteString("<text x=\"32\" y=\"40\" fill=\"#00eaff\" font-family=\"monospace\" font-size=\"24\">RESTLESS topology fallback</text>")
	for i, node := range nodes {
		y := 80 + i*38
		svg.WriteString(fmt.Sprintf("<circle cx=\"48\" cy=\"%d\" r=\"6\" fill=\"#8a2be2\"/>", y-6))
		svg.WriteString(fmt.Sprintf("<text x=\"64\" y=\"%d\" fill=\"#e6edf3\" font-family=\"monospace\" font-size=\"18\">%s</text>", y, escapeXML(node)))
	}
	svg.WriteString("</svg>")
	return os.WriteFile(file, []byte(svg.String()), 0o644)
}

func cleanNode(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ";")
	s = strings.Trim(s, "\"")
	return s
}

func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
	)
	return replacer.Replace(s)
}
