package topology

import "strings"

func Build(host string, paths []string) string {

	tree := host + "\n"

	for _, p := range paths {

		p = strings.Trim(p, "/")

		if p == "" {
			continue
		}

		tree += " ├── " + p + "\n"
	}

	return tree
}
