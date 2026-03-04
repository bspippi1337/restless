package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bspippi1337/restless/internal/core/state"
	"github.com/spf13/cobra"
)

type node struct {
	Name     string
	Children map[string]*node
}

func newNode(name string) *node {
	return &node{
		Name:     name,
		Children: map[string]*node{},
	}
}

func buildTree(paths []string) *node {
	root := newNode("/")

	for _, p := range paths {

		p = strings.TrimPrefix(p, "/")

		parts := strings.Split(p, "/")

		cur := root

		for _, part := range parts {

			if part == "" {
				continue
			}

			if cur.Children[part] == nil {
				cur.Children[part] = newNode(part)
			}

			cur = cur.Children[part]
		}
	}

	return root
}

func printTree(n *node, prefix string) {

	keys := make([]string, 0, len(n.Children))

	for k := range n.Children {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for i, k := range keys {

		child := n.Children[k]

		connector := "├─"
		nextPrefix := prefix + "│ "

		if i == len(keys)-1 {
			connector = "└─"
			nextPrefix = prefix + "  "
		}

		fmt.Printf("%s%s %s\n", prefix, connector, child.Name)

		printTree(child, nextPrefix)
	}
}

func NewGraphCmd() *cobra.Command {

	return &cobra.Command{
		Use:   "graph",
		Short: "Render API topology graph",
		RunE: func(cmd *cobra.Command, args []string) error {

			st, _, err := state.Load()
			if err != nil {
				return err
			}

			var paths []string

			for _, r := range st.LastScan.Endpoints {
				paths = append(paths, r.Path)
			}

			if len(paths) == 0 {
				fmt.Println("No endpoints in state. Run:")
				fmt.Println("  restless discover <url>")
				return nil
			}

			tree := buildTree(paths)

			fmt.Println("/")
			printTree(tree, "")

			return nil
		},
	}
}
