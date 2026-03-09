package engine

import "strings"

func nodeStyle(name string) string {

	if name == "root" {
		return `shape=oval,style=filled,fillcolor="#eeeeee"`
	}

	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		return `shape=diamond,style=filled,fillcolor="#ffe8cc"`
	}

	return `shape=box,style=rounded,style=filled,fillcolor="#e3f2fd"`
}
