package magiswarm

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

func BuildTopologyDOT(host string, eps []Endpoint) string {

	var b strings.Builder

	b.WriteString("digraph restless {\n")
	b.WriteString("rankdir=LR;\n")
	b.WriteString("node [shape=box,style=rounded,fontname=monospace];\n\n")

	b.WriteString("\"" + host + "\" [shape=oval];\n")

	for _, e := range eps {

		p := strings.Trim(e.Path, "/")
		if p == "" {
			continue
		}

		parent := host
		parts := strings.Split(p, "/")

		for _, seg := range parts {

			cur := parent + "/" + seg

			b.WriteString("\"" + parent + "\" -> \"" + cur + "\";\n")

			parent = cur
		}
	}

	b.WriteString("}\n")

	return b.String()
}

func TryRenderSVG(dotfile string, svgfile string) {

	if _, err := exec.LookPath("dot"); err != nil {
		return
	}

	data, err := os.ReadFile(dotfile)
	if err != nil {
		return
	}

	cmd := exec.Command("dot", "-Tsvg")

	cmd.Stdin = bytes.NewReader(data)

	out, err := cmd.Output()
	if err != nil {
		return
	}

	os.WriteFile(svgfile, out, 0644)
}
