#!/usr/bin/env bash

cd "$(dirname "$0")/.."

echo "== installing endpoint sorter =="

cat > internal/engine/sort.go <<'GO'
package engine

import "sort"

func SortEndpoints(e []Endpoint) []Endpoint {

	sort.Slice(e, func(i,j int) bool {
		return e[i].Path < e[j].Path
	})

	return e
}
GO


echo "== installing smarter resource classifier =="

cat > internal/engine/classify.go <<'GO'
package engine

import "strings"

func classifyConfidence(p string) string {

	if strings.Contains(p,"{") {
		return "high"
	}

	if strings.Count(p,"/") <= 1 {
		return "medium"
	}

	return "medium"
}
GO


echo "== upgrading cleanup pipeline =="

cat > internal/engine/cleanup.go <<'GO'
package engine

func CleanEndpoints(e []Endpoint) []Endpoint {

	tmp := []Endpoint{}

	for _,ep := range e {

		if !ValidEndpoint(ep.Path) {
			continue
		}

		ep.Path = NormalizeTemplate(ep.Path)
		ep.Confidence = classifyConfidence(ep.Path)

		tmp = append(tmp,ep)
	}

	tmp = Deduplicate(tmp)
	tmp = SortEndpoints(tmp)

	return tmp
}
GO


echo "== formatting =="
go fmt ./...

echo "== rebuilding =="
rm -rf build
make build

echo
echo "================================="
echo "Restless engine polished"
echo "================================="
echo
echo "Run:"
echo "./build/restless api.github.com"
