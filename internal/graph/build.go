package graph

import "github.com/bspippi1337/restless/internal/core"

type Node struct {
	Path    string
	Methods []string
}

func Build(endpoints []core.Endpoint) []Node {
	m := map[string][]string{}

	for _, ep := range endpoints {
		m[ep.Path] = append(m[ep.Path], ep.Method)
	}

	var nodes []Node

	for p, methods := range m {
		nodes = append(nodes, Node{
			Path:    p,
			Methods: methods,
		})
	}

	return nodes
}
