package graph

import (
	"fmt"
	"io"
)

func RenderSVG(w io.Writer, nodes []Node) {

	width := 600
	height := 80 + len(nodes)*60

	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`+"\n", width, height)

	y := 40

	for _, n := range nodes {

		fmt.Fprintf(w,
			`<text x="20" y="%d" font-family="monospace" font-size="16">%s</text>`+"\n",
			y,
			n.Path,
		)

		my := y + 18

		for _, m := range n.Methods {

			fmt.Fprintf(w,
				`<text x="40" y="%d" font-family="monospace" font-size="14">%s</text>`+"\n",
				my,
				m,
			)

			my += 16
		}

		y += 60
	}

	fmt.Fprintln(w, "</svg>")
}
