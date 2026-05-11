package graph

import (
	"fmt"
	"io"
)

type Node struct {
	Path    string
	Methods []string
}

func RenderASCII(w io.Writer, nodes []Node) {
	fmt.Fprintln(w, "API MAP")
	fmt.Fprintln(w)
	for _, n := range nodes {
		fmt.Fprintf(w, "%s\n", n.Path)
		for _, m := range n.Methods {
			fmt.Fprintf(w, "  %s\n", m)
		}
		fmt.Fprintln(w)
	}
}
