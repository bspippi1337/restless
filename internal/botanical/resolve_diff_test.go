package botanical

import (
	"testing"
)

func TestResolveLocalOpenAPIParameters(t *testing.T) {
	known := map[string]Parameter{
		"PetId":      {Name: "petId", In: "path", Required: true},
		"Cursor":     {Name: "cursor", In: "query"},
		"Alias":      {Ref: "#/components/parameters/PetId"},
		"LoopA":      {Ref: "#/components/parameters/LoopB"},
		"LoopB":      {Ref: "#/components/parameters/LoopA"},
		"Slash/Name": {Name: "escaped", In: "header"},
	}
	for _, tc := range []struct {
		name       string
		input      Parameter
		unresolved bool
		wantName   string
	}{
		{"direct", Parameter{Ref: "#/components/parameters/PetId"}, false, "petId"},
		{"recursive local", Parameter{Ref: "#/components/parameters/Alias"}, false, "petId"},
		{"json-pointer escape", Parameter{Ref: "#/components/parameters/Slash~1Name"}, false, "escaped"},
		{"missing", Parameter{Ref: "#/components/parameters/Missing"}, true, ""},
		{"external blocked", Parameter{Ref: "https://evil.example/parameters/Secret"}, true, ""},
		{"cyclic", Parameter{Ref: "#/components/parameters/LoopA"}, true, ""},
		{"malformed", Parameter{Ref: "#/components/parameters/Bad~2Escape"}, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolved, bad := resolveParameters([]Parameter{tc.input}, known)
			if bad != tc.unresolved || len(resolved) != 1 || resolved[0].Name != tc.wantName {
				t.Fatalf("got params=%+v unresolved=%v", resolved, bad)
			}
		})
	}
}

func TestCatalogExpandsLocalComponentsWithoutNetwork(t *testing.T) {
	spec := `{"openapi":"3.0.4","info":{"title":"Local refs"},"components":{"parameters":{"ID":{"name":"id","in":"path","required":true,"schema":{"type":"string"}}}},"paths":{"/items/{id}":{"parameters":[{"$ref":"#/components/parameters/ID"}],"get":{"operationId":"getItem","summary":"Read item"}}}}`
	dir := fixture(t, spec, "application/json")
	c, err := BuildCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Operations) != 1 {
		t.Fatalf("got %d operations", len(c.Operations))
	}
	op := c.Operations[0]
	if op.UnresolvedRefs || len(op.Parameters) != 1 || op.Parameters[0].Name != "id" || !op.Parameters[0].Required {
		t.Fatalf("local parameter reference did not resolve: %+v", op)
	}
	plan, err := BuildProbePlan(c, "read item", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Candidates) != 1 || !plan.Candidates[0].RequiresManualReview || plan.NetworkRequests != 0 {
		t.Fatalf("unexpected probe plan: %+v", plan)
	}
}

func TestOfflineDiffTracksContractChangeNotDocumentBytes(t *testing.T) {
	initial := Catalog{Version: catalogVersion, Operations: []Operation{
		{SourceURL: "https://api.example/spec.json", SourceSHA256: "old", Method: "GET", Path: "/items", OperationID: "listItems", Summary: "List items"},
		{SourceURL: "https://api.example/spec.json", SourceSHA256: "old", Method: "POST", Path: "/items", OperationID: "createItem"},
	}}
	next := Catalog{Version: catalogVersion, Operations: []Operation{
		{SourceURL: "https://api.example/spec.json", SourceSHA256: "new", Method: "GET", Path: "/items", OperationID: "listItems", Summary: "List items"},
		{SourceURL: "https://api.example/spec.json", SourceSHA256: "new", Method: "PATCH", Path: "/items/{id}", OperationID: "patchItem"},
	}}
	diff, err := CompareCatalogs(initial, next)
	if err != nil {
		t.Fatal(err)
	}
	if diff.NetworkRequests != 0 || diff.Unchanged != 1 || len(diff.Added) != 1 || len(diff.Removed) != 1 || len(diff.Changed) != 0 {
		t.Fatalf("incorrect diff: %+v", diff)
	}
	next.Operations[0].Summary = "List all items including archived"
	diff, err = CompareCatalogs(initial, next)
	if err != nil {
		t.Fatal(err)
	}
	if diff.Unchanged != 0 || len(diff.Changed) != 1 || diff.Changed[0].Before == nil || diff.Changed[0].After == nil {
		t.Fatalf("summary change not detected: %+v", diff)
	}
	next.Operations = append(next.Operations, next.Operations[0])
	if _, err := CompareCatalogs(initial, next); err == nil {
		t.Fatal("duplicate operation keys not rejected")
	}
}
