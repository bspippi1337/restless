package topology

import (
	"sort"
	"strings"
)

type node struct {
	children map[string]*node
}

func newNode() *node { return &node{children: map[string]*node{}} }

func (n *node) addPath(p string) {
	p = strings.TrimLeft(p, "/")
	if p == "" {
		return
	}
	cur := n
	for _, seg := range strings.Split(p, "/") {
		if seg == "" {
			continue
		}
		ch, ok := cur.children[seg]
		if !ok {
			ch = newNode()
			cur.children[seg] = ch
		}
		cur = ch
	}
}

func ASCII(host string, paths []string) string {
	root := newNode()
	for _, p := range paths {
		root.addPath(p)
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
			last := i == len(keys)-1
			branch := "├── "
			next := prefix + "│   "
			if last {
				branch = "└── "
				next = prefix + "    "
			}
			sb.WriteString(prefix)
			sb.WriteString(branch)
			sb.WriteString(k)
			sb.WriteString("\n")
			walk(n.children[k], next)
		}
	}
	walk(root, "")
	return sb.String()
}
