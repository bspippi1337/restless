package graph

import (
	"fmt"
	"io"
)

func RenderSVG(w io.Writer, nodes []Node) {

	height := len(nodes)*70 + 40
	width := 600

	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`,
		width, height, width, height)

	fmt.Fprint(w, `<rect width="100%" height="100%" fill="white"/>`)

	y := 40

	for _, n := range nodes {

		fmt.Fprintf(
			w,
			`<text x="20" y="%d" font-family="monospace" font-size="16" fill="black">%s</text>`,
			y,
			n.Path,
		)

		my := y + 20

		for _, m := range n.Methods {

			fmt.Fprintf(
				w,
				`<text x="40" y="%d" font-family="monospace" font-size="14" fill="black">%s</text>`,
				my,
				m,
			)

			my += 18
		}

		y += 70
	}

	fmt.Fprint(w, `</svg>`)
}
