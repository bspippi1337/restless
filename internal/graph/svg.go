package graph

import (
	"fmt"
	"io"
)

func RenderSVG(w io.Writer, nodes []Node) {

	height := len(nodes)*60 + 40

	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="%d">`, height)

	y := 40

	for _, n := range nodes {

		fmt.Fprintf(w,
			`<text x="20" y="%d" font-family="monospace" font-size="16">%s</text>`,
			y,
			n.Path,
		)

		my := y + 18

		for _, m := range n.Methods {

			fmt.Fprintf(w,
				`<text x="40" y="%d" font-family="monospace" font-size="14">%s</text>`,
				my,
				m,
			)

			my += 16
		}

		y += 60
	}

	fmt.Fprint(w, `</svg>`)
}
