package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
)

type Validator struct {
	doc *openapi3.T
}

func NewValidator(doc *openapi3.T) *Validator { return &Validator{doc: doc} }

func (v *Validator) ValidateResponse(ctx context.Context, method, pathTemplate string, status int, contentType string, body []byte) ([]model.Finding, error) {
	op, opID, err := findOperation(v.doc, method, pathTemplate)
	if err != nil {
		return nil, err
	}

	rr := pickResponse(op, status)
	if rr == nil || rr.Value == nil {
		return []model.Finding{{
			OpID: opID, Method: strings.ToUpper(method), Path: pathTemplate, Status: status, ContentType: contentType,
			Kind: model.KindSchemaViolation, Severity: model.SevMedium,
			JSONPath: "$", Message: "response not defined in OpenAPI spec",
		}}, nil
	}

	mt := pickMediaType(rr.Value, contentType)
	if mt == nil || mt.Schema == nil || mt.Schema.Value == nil {
		return []model.Finding{{
			OpID: opID, Method: strings.ToUpper(method), Path: pathTemplate, Status: status, ContentType: contentType,
			Kind: model.KindSchemaViolation, Severity: model.SevMedium,
			JSONPath: "$", Message: "response schema missing for content-type",
		}}, nil
	}

	js, err := json.Marshal(mt.Schema.Value)
	if err != nil {
		return nil, err
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(js)); err != nil {
		return nil, err
	}
	sch, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, err
	}

	var payload any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return []model.Finding{{
			OpID: opID, Method: strings.ToUpper(method), Path: pathTemplate, Status: status, ContentType: contentType,
			Kind: model.KindSchemaViolation, Severity: model.SevHigh,
			JSONPath: "$", Message: "response body is not valid JSON",
			Actual: err.Error(),
		}}, nil
	}

	if err := sch.Validate(payload); err != nil {
		return mapSchemaError(opID, method, pathTemplate, status, contentType, err), nil
	}

	return nil, nil
}

func pickMediaType(resp *openapi3.Response, contentType string) *openapi3.MediaType {
	if resp == nil || resp.Content == nil {
		return nil
	}
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if ct == "" {
		ct = "application/json"
	}
	if mt := resp.Content.Get(ct); mt != nil {
		return mt
	}
	if mt := resp.Content.Get("application/json"); mt != nil {
		return mt
	}
	for k := range resp.Content {
		if strings.Contains(k, "json") {
			return resp.Content.Get(k)
		}
	}
	return nil
}

func pickResponse(op *openapi3.Operation, status int) *openapi3.ResponseRef {
	if op == nil || op.Responses == nil {
		return nil
	}
	code := fmt.Sprintf("%d", status)
	if rr := op.Responses.Get(code); rr != nil {
		return rr
	}
	class := fmt.Sprintf("%dXX", status/100)
	if rr := op.Responses.Get(class); rr != nil {
		return rr
	}
	if rr := op.Responses.Default(); rr != nil {
		return rr
	}
	return nil
}

func findOperation(doc *openapi3.T, method, pathTemplate string) (*openapi3.Operation, string, error) {
	item := doc.Paths.Find(pathTemplate)
	if item == nil {
		return nil, "", fmt.Errorf("path not found in spec: %s", pathTemplate)
	}
	m := strings.ToUpper(method)
	var op *openapi3.Operation
	switch m {
	case http.MethodGet:
		op = item.Get
	case http.MethodPost:
		op = item.Post
	case http.MethodPut:
		op = item.Put
	case http.MethodPatch:
		op = item.Patch
	case http.MethodDelete:
		op = item.Delete
	case http.MethodHead:
		op = item.Head
	case http.MethodOptions:
		op = item.Options
	case http.MethodTrace:
		op = item.Trace
	default:
		return nil, "", fmt.Errorf("unsupported method: %s", m)
	}
	if op == nil {
		return nil, "", fmt.Errorf("operation missing in spec: %s %s", m, pathTemplate)
	}
	opID := op.OperationID
	if opID == "" {
		opID = strings.ToLower(m) + " " + pathTemplate
	}
	return op, opID, nil
}
