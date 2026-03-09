package engine

import (
	"strings"
)

func BuildTopology(endpoints []string) string {

	tree := "root\n"

	for _, e := range endpoints {

		parts := strings.Split(strings.TrimPrefix(e, "/"), "/")

		for i, p := range parts {

			prefix := strings.Repeat("  ", i+1)
			tree += prefix + "└── " + p + "\n"

		}
	}

	return tree
}
