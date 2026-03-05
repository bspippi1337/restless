package recon

import (
	"context"
	"encoding/json"
	"net/url"
)

type GraphQLInfo struct {
	Endpoint      string `json:"endpoint"`
	Introspection bool   `json:"introspection"`
	Types         int    `json:"types"`
	Note          string `json:"note"`
}

// TryGraphQLIntrospection performs a tiny schema introspection (bounded) on /graphql.
func TryGraphQLIntrospection(ctx context.Context, e *Engine, base *url.URL) (*GraphQLInfo, error) {
	u := *base
	u.Path = "/graphql"

	query := `{"query":"query IntrospectionQuery{__schema{types{name}}}"}`
	resp, err := e.Request(ctx, "POST", u.String(), []byte(query))
	if err != nil {
		return nil, err
	}
	gi := &GraphQLInfo{Endpoint: u.Path}
	if resp.Status >= 500 {
		gi.Note = "server-error"
		return gi, nil
	}
	if resp.Status == 404 {
		gi.Note = "not-found"
		return gi, nil
	}
	if resp.Status == 401 || resp.Status == 403 {
		gi.Note = "auth-required"
		return gi, nil
	}
	if !LooksJSON(resp.ContentType, resp.Body) {
		gi.Note = "non-json"
		return gi, nil
	}
	var doc map[string]any
	if err := json.Unmarshal(resp.Body, &doc); err != nil {
		gi.Note = "json-parse-failed"
		return gi, nil
	}
	data, ok := doc["data"].(map[string]any)
	if !ok {
		gi.Note = "no-data"
		return gi, nil
	}
	schema, ok := data["__schema"].(map[string]any)
	if !ok {
		gi.Note = "no-schema"
		return gi, nil
	}
	if types, ok := schema["types"].([]any); ok {
		gi.Types = len(types)
	}
	gi.Introspection = gi.Types > 0
	gi.Note = "ok"
	return gi, nil
}
