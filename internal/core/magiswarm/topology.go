package magiswarm

import (
	"sort"
	"strings"
)

type node struct {
	name     string
	children map[string]*node
}

func newNode(name string) *node {
	return &node{name: name, children: map[string]*node{}}
}

func (n *node) add(parts []string) {
	cur := n
	for _, p := range parts {
		if p == "" {
			continue
		}
		ch, ok := cur.children[p]
		if !ok {
			ch = newNode(p)
			cur.children[p] = ch
		}
		cur = ch
	}
}

func BuildTopologyASCII(host string, eps []Endpoint) string {
	root := newNode(host)

	for _, e := range eps {
		p := strings.TrimSpace(e.Path)
		if p == "" {
			continue
		}
		p = strings.TrimLeft(p, "/")
		parts := strings.Split(p, "/")
		if p == "" {
			parts = []string{"/"}
		}
		root.add(parts)
	}

	var sb strings.Builder
	sb.WriteString(host)
	sb.WriteString("\n")

	var walk func(n *node, prefix string)
	walk = func(n *node, prefix string) {
		keys := make([]string, 0, len(n.children))
		for k := range n.children {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			ch := n.children[k]
			isLast := i == len(keys)-1

			branch := "├── "
			nextPrefix := prefix + "│   "
			if isLast {
				branch = "└── "
				nextPrefix = prefix + "    "
			}

			sb.WriteString(prefix)
			sb.WriteString(branch)
			sb.WriteString(k)
			sb.WriteString("\n")

			walk(ch, nextPrefix)
		}
	}
	walk(root, "")
	return sb.String()
}
