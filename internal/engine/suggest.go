package engine

import "github.com/bspippi1337/restless/internal/discover"

func Suggest(p discover.Profile) []string {
	out := []string{"request-builder"}
	if len(p.ContentTypes) > 0 {
		out = append(out, "export")
	}
	if len(p.Methods) > 0 {
		out = append(out, "fuzzer")
	}
	return out
}
