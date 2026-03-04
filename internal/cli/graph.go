package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

type node struct {
	name     string
	children map[string]*node
}

func newNode(n string) *node {
	return &node{name: n, children: map[string]*node{}}
}

func build(paths []string) *node {

	root := newNode("/")

	for _, p := range paths {

		p = strings.TrimPrefix(p, "/")
		parts := strings.Split(p, "/")

		cur := root

		for _, x := range parts {

			if x == "" {
				continue
			}

			if cur.children[x] == nil {
				cur.children[x] = newNode(x)
			}

			cur = cur.children[x]
		}
	}

	return root
}

func printTree(n *node, prefix string) {

	keys := make([]string, 0, len(n.children))

	for k := range n.children {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for i, k := range keys {

		child := n.children[k]

		conn := "├─"
		next := prefix + "│ "

		if i == len(keys)-1 {
			conn = "└─"
			next = prefix + "  "
		}

		fmt.Printf("%s%s %s\n", prefix, conn, child.name)

		printTree(child, next)
	}
}

func writeSVG(path string) {

	os.MkdirAll("dist", 0755)

	f, _ := os.Create(path)
	defer f.Close()

	fmt.Fprintln(f, `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="200">`)
	fmt.Fprintln(f, `<text x="20" y="20" font-family="monospace">Restless API Graph</text>`)
	fmt.Fprintln(f, `</svg>`)
}

func NewGraphCmd() *cobra.Command {

	var svg bool

	cmd := &cobra.Command{

		Use: "graph",

		RunE: func(cmd *cobra.Command, args []string) error {

			st, _, err := state.Load()
			if err != nil {
				return err
			}

			var paths []string

			for _, r := range st.LastScan.Endpoints {
				paths = append(paths, r.Path)
			}

			root := build(paths)

			fmt.Println("/")
			printTree(root, "")

			if svg {
				writeSVG("dist/api-topology.svg")
				fmt.Println("SVG written to dist/api-topology.svg")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&svg, "svg", false, "export svg")

	return cmd
}