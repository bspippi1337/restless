package ascii

import (
	"fmt"
	"sort"
	"strings"
)

type Node struct {
	Name     string
	Children map[string]*Node
}

func NewNode(name string) *Node {
	return &Node{Name: name, Children: map[string]*Node{}}
}

func BuildTree(paths []string) *Node {
	root := NewNode("/")

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = strings.TrimPrefix(p, "/")
		parts := strings.Split(p, "/")

		cur := root
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if _, ok := cur.Children[part]; !ok {
				cur.Children[part] = NewNode(part)
			}
			cur = cur.Children[part]
		}
	}

	return root
}

func Render(root *Node) {
	fmt.Println("/")

	keys := make([]string, 0, len(root.Children))
	for k := range root.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		renderNode(root.Children[k], "", i == len(keys)-1)
	}
}

func renderNode(n *Node, prefix string, last bool) {
	connector := "  "
	nextPrefix := prefix + "  "
	if last {
		connector = "  "
		nextPrefix = prefix + "  "
	}

	fmt.Println(prefix + connector + n.Name)

	keys := make([]string, 0, len(n.Children))
	for k := range n.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		renderNode(n.Children[k], nextPrefix, i == len(keys)-1)
	}
}
