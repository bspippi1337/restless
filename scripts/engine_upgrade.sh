#!/usr/bin/env bash

cd "$(dirname "$0")/.."

echo "== installing endpoint template inference =="

cat > internal/engine/template.go <<'GO'
package engine

import (
	"regexp"
	"strings"
)

var num = regexp.MustCompile(`^[0-9]+$`)
var hex = regexp.MustCompile(`^[0-9a-fA-F]{6,}$`)
var uuid = regexp.MustCompile(`^[0-9a-fA-F\-]{32,36}$`)

func NormalizeTemplate(p string) string {

	parts := strings.Split(strings.Trim(p,"/"),"/")

	for i,v := range parts {

		if num.MatchString(v) {
			parts[i] = "{id}"
			continue
		}

		if uuid.MatchString(v) {
			parts[i] = "{uuid}"
			continue
		}

		if hex.MatchString(v) {
			parts[i] = "{hash}"
			continue
		}

		if len(v) > 12 {
			parts[i] = "{value}"
		}
	}

	return "/" + strings.Join(parts,"/")
}
GO


echo "== installing endpoint deduplicator =="

cat > internal/engine/dedupe.go <<'GO'
package engine

func Deduplicate(endpoints []Endpoint) []Endpoint {

	seen := map[string]bool{}
	out := []Endpoint{}

	for _,e := range endpoints {

		t := NormalizeTemplate(e.Path)

		if seen[t] {
			continue
		}

		seen[t] = true

		e.Path = t
		out = append(out,e)
	}

	return out
}
GO


echo "== installing discovery filter =="

cat > internal/engine/filter.go <<'GO'
package engine

import "strings"

func ValidEndpoint(p string) bool {

	if p == "" {
		return false
	}

	if !strings.HasPrefix(p,"/") {
		return false
	}

	if strings.Contains(p," ") {
		return false
	}

	if strings.Count(p,"/") > 6 {
		return false
	}

	return true
}
GO


echo "== patching engine pipeline =="

cat > internal/engine/pipeline_patch.go <<'GO'
package engine

func CleanEndpoints(e []Endpoint) []Endpoint {

	tmp := []Endpoint{}

	for _,ep := range e {

		if !ValidEndpoint(ep.Path) {
			continue
		}

		tmp = append(tmp,ep)
	}

	tmp = Deduplicate(tmp)

	return tmp
}
GO


echo "== wiring cleaner into engine =="

cat <<'GO' >> internal/engine/engine.go

// automatic cleanup stage
func cleanupEndpoints(e []Endpoint) []Endpoint {
	return CleanEndpoints(e)
}
GO


echo "== formatting =="
go fmt ./...

echo "== rebuilding =="
rm -rf build
make build

echo
echo "=================================="
echo "Restless discovery engine upgraded"
echo "=================================="
echo
echo "Run:"
echo "./build/restless api.github.com"
