package diff

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
)

type opKey struct {
	Method string
	Path   string
}

func Diff(ctx context.Context, oldDoc, newDoc *openapi3.T) (*model.DiffResult, error) {
	_ = ctx
	oldOps := index(oldDoc)
	newOps := index(newDoc)

	var breaking, nonBreaking []string

	for k := range oldOps {
		if _, ok := newOps[k]; !ok {
			breaking = append(breaking, fmt.Sprintf("removed operation: %s %s", k.Method, k.Path))
		}
	}
	for k := range newOps {
		if _, ok := oldOps[k]; !ok {
			nonBreaking = append(nonBreaking, fmt.Sprintf("added operation: %s %s", k.Method, k.Path))
		}
	}

	for k, oldOp := range oldOps {
		newOp, ok := newOps[k]
		if !ok {
			continue
		}
		breaking = append(breaking, compareResponses(k, oldOp, newOp)...)
		nonBreaking = append(nonBreaking, compareResponsesNonBreaking(k, oldOp, newOp)...)
	}

	sort.Strings(breaking)
	sort.Strings(nonBreaking)

	return &model.DiffResult{
		Breaking:        breaking,
		NonBreaking:     nonBreaking,
		RecommendedBump: recommend(breaking, nonBreaking),
	}, nil
}

func index(doc *openapi3.T) map[opKey]*openapi3.Operation {
	out := map[opKey]*openapi3.Operation{}
	if doc == nil || doc.Paths == nil {
		return out
	}
	for p, it := range doc.Paths.Map() {
		if it == nil {
			continue
		}
		put := func(m string, op *openapi3.Operation) {
			if op != nil {
				out[opKey{Method: strings.ToUpper(m), Path: p}] = op
			}
		}
		put("GET", it.Get)
		put("POST", it.Post)
		put("PUT", it.Put)
		put("PATCH", it.Patch)
		put("DELETE", it.Delete)
		put("HEAD", it.Head)
		put("OPTIONS", it.Options)
		put("TRACE", it.Trace)
	}
	return out
}

func compareResponses(k opKey, oldOp, newOp *openapi3.Operation) []string {
	var breaking []string
	oldR := respCodes(oldOp)
	newR := respCodes(newOp)

	for code := range oldR {
		if _, ok := newR[code]; !ok {
			breaking = append(breaking, fmt.Sprintf("%s %s: removed response %s", k.Method, k.Path, code))
		}
	}
	for code := range oldR {
		if _, ok := newR[code]; !ok {
			continue
		}
		ot := schemaType(oldOp, code)
		nt := schemaType(newOp, code)
		if ot != "" && nt != "" && ot != nt {
			breaking = append(breaking, fmt.Sprintf("%s %s: response %s type changed (%s -> %s)", k.Method, k.Path, code, ot, nt))
		}
	}
	return breaking
}

func compareResponsesNonBreaking(k opKey, oldOp, newOp *openapi3.Operation) []string {
	var nb []string
	oldR := respCodes(oldOp)
	newR := respCodes(newOp)
	for code := range newR {
		if _, ok := oldR[code]; !ok {
			nb = append(nb, fmt.Sprintf("%s %s: added response %s", k.Method, k.Path, code))
		}
	}
	return nb
}

func respCodes(op *openapi3.Operation) map[string]struct{} {
	out := map[string]struct{}{}
	if op == nil || op.Responses == nil {
		return out
	}
	for c := range op.Responses.Map() {
		out[c] = struct{}{}
	}
	if op.Responses.Default() != nil {
		out["default"] = struct{}{}
	}
	return out
}

func schemaType(op *openapi3.Operation, code string) string {
	if op == nil || op.Responses == nil {
		return ""
	}
	var rr *openapi3.ResponseRef
	if code == "default" {
		rr = op.Responses.Default()
	} else {
		rr = op.Responses.Map()[code]
	}
	if rr == nil || rr.Value == nil || rr.Value.Content == nil {
		return ""
	}
	mt := rr.Value.Content.Get("application/json")
	if mt == nil || mt.Schema == nil || mt.Schema.Value == nil {
		for k := range rr.Value.Content {
			if strings.Contains(k, "json") {
				mt = rr.Value.Content.Get(k)
				break
			}
		}
	}
	if mt == nil || mt.Schema == nil || mt.Schema.Value == nil {
		return ""
	}
	if mt.Schema.Value.Type != nil && len(*mt.Schema.Value.Type) > 0 { return (*mt.Schema.Value.Type)[0] }
	return ""
}

func recommend(breaking, nonBreaking []string) model.SemverBump {
	if len(breaking) > 0 {
		return model.BumpMajor
	}
	if len(nonBreaking) > 0 {
		return model.BumpMinor
	}
	return model.BumpPatch
}
