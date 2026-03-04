package topology

import (
	"fmt"
	"sort"
	"strings"
)

// Node represents an API route segment in a tree.
type Node struct {
	Name     string
	Children map[string]*Node
	Methods  map[string]bool
}

func NewNode(name string) *Node {
	return &Node{
		Name:     name,
		Children: map[string]*Node{},
		Methods:  map[string]bool{},
	}
}

// BuildTree builds a route tree from discovered endpoints like:
// GET /v1/users
// POST /v1/login
func BuildTree(lines []string) *Node {
	root := NewNode("API")

	for _, ln := range lines {
		parts := strings.Fields(strings.TrimSpace(ln))
		if len(parts) < 2 {
			continue
		}
		method := strings.ToUpper(parts[0])
		path := parts[1]

		segments := strings.Split(strings.Trim(path, "/"), "/")

		node := root
		for _, seg := range segments {
			if seg == "" {
				continue
			}
			if node.Children[seg] == nil {
				node.Children[seg] = NewNode(seg)
			}
			node = node.Children[seg]
		}

		node.Methods[method] = true
	}

	return root
}

// Render prints the topology as ASCII tree.
func Render(n *Node, prefix string, last bool) {

	connector := "├─"
	nextPrefix := prefix + "│ "

	if last {
		connector = "└─"
		nextPrefix = prefix + "  "
	}

	if prefix == "" {
		fmt.Println(n.Name)
	} else {
		methods := ""
		if len(n.Methods) > 0 {
			var m []string
			for k := range n.Methods {
				m = append(m, k)
			}
			sort.Strings(m)
			methods = " [" + strings.Join(m, ",") + "]"
		}
		fmt.Printf("%s%s %s%s\n", prefix, connector, n.Name, methods)
	}

	keys := make([]string, 0, len(n.Children))
	for k := range n.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		Render(n.Children[k], nextPrefix, i == len(keys)-1)
	}
}
