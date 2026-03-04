package ascii

import (
	"fmt"
	"os"
	"sort"
)

func WriteSVG(root *Node, path string) error {

	lines := []string{}
	collect(root, "", &lines)

	width := 800
	height := 20 * (len(lines) + 2)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" font-family="monospace" font-size="14">`, width, height)

	y := 20
	for _, l := range lines {
		fmt.Fprintf(f, `<text x="10" y="%d">%s</text>`, y, l)
		y += 20
	}

	fmt.Fprintln(f, "</svg>")
	return nil
}

func collect(n *Node, prefix string, out *[]string) {

	if n.Name == "/" {
		*out = append(*out, "/")
	} else {
		*out = append(*out, prefix+n.Name)
	}

	keys := make([]string, 0, len(n.Children))
	for k := range n.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		collect(n.Children[k], prefix+"  ", out)
	}
}
