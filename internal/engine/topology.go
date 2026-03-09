package engine

import (
	"sort"
	"strings"
)

type node struct {
	children map[string]*node
}

func newNode() *node {
	return &node{children: map[string]*node{}}
}

func BuildTopology(endpoints []string) string {

	root := newNode()

	for _, ep := range endpoints {

		clean := strings.Trim(ep, "/")
		if clean == "" {
			continue
		}

		parts := strings.Split(clean, "/")

		cur := root

		for _, p := range parts {

			if _, ok := cur.children[p]; !ok {
				cur.children[p] = newNode()
			}

			cur = cur.children[p]
		}
	}

	var b strings.Builder
	b.WriteString("root\n")

	render(&b, root, 1)

	return b.String()
}

func render(b *strings.Builder, n *node, depth int) {

	keys := make([]string, 0, len(n.children))

	for k := range n.children {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {

		b.WriteString(strings.Repeat("  ", depth))
		b.WriteString("└── ")
		b.WriteString(k)
		b.WriteString("\n")

		render(b, n.children[k], depth+1)
	}
}
