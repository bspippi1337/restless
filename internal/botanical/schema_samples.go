package botanical

import (
	"fmt"
	"sort"
	"strings"
)

// RequestSample is a bounded, explicitly documented example only. It does not
// assert that the endpoint is live, accepts the example, or is safe to call.
type RequestSample struct {
	Required      bool   `json:"required"`
	ContentType   string `json:"content_type,omitempty"`
	Example       any    `json:"example,omitempty"`
	UnresolvedRef string `json:"unresolved_ref,omitempty"`
}

type ResponseSummary struct {
	Status        string `json:"status"`
	Description   string `json:"description,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	Example       any    `json:"example,omitempty"`
	UnresolvedRef string `json:"unresolved_ref,omitempty"`
}

type rawSchema struct {
	Type       string               `yaml:"type"`
	Format     string               `yaml:"format"`
	Ref        string               `yaml:"$ref"`
	Example    any                  `yaml:"example"`
	Default    any                  `yaml:"default"`
	Enum       []any                `yaml:"enum"`
	Properties map[string]rawSchema `yaml:"properties"`
	Items      *rawSchema           `yaml:"items"`
}

type rawMediaType struct {
	Schema  rawSchema `yaml:"schema"`
	Example any       `yaml:"example"`
}

type rawRequestBody struct {
	Ref      string                  `yaml:"$ref"`
	Required bool                    `yaml:"required"`
	Content  map[string]rawMediaType `yaml:"content"`
}

type rawResponse struct {
	Ref         string                  `yaml:"$ref"`
	Description string                  `yaml:"description"`
	Content     map[string]rawMediaType `yaml:"content"`
}

// Schema previews are deliberately small, avoid alias expansion, and never
// dereference even a local $ref: leave that provenance ambiguity visible.
func boundedExample(value any, depth int) any {
	if depth > 4 {
		return "<nested value>"
	}
	switch x := value.(type) {
	case nil:
		return nil
	case string:
		if len(x) > 200 {
			return x[:200] + "…"
		}
		return x
	case bool, int, int64, float64, float32, uint64:
		return x
	case map[string]any:
		keys := make([]string, 0, len(x))
		for key := range x {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out := make(map[string]any)
		for i, key := range keys {
			if i >= 20 {
				out["__truncated__"] = "maximum 20 example properties"
				break
			}
			if redactedField(key) {
				out[key] = "<REDACTED: enter locally>"
				continue
			}
			out[key] = boundedExample(x[key], depth+1)
		}
		return out
	case []any:
		out := make([]any, 0, 1)
		if len(x) > 0 {
			out = append(out, boundedExample(x[0], depth+1))
		}
		return out
	default:
		return fmt.Sprint(x)
	}
}
func redactedField(name string) bool {
	v := strings.ToLower(name)
	for _, token := range []string{"password", "passwd", "secret", "api_key", "apikey", "access_token", "refresh_token", "private_key"} {
		if strings.Contains(v, token) {
			return true
		}
	}
	return false
}
func schemaExample(schema rawSchema, name string, depth int) any {
	if redactedField(name) {
		return "<REDACTED: enter locally>"
	}
	if depth >= 4 {
		return "<nested schema>"
	}
	if schema.Ref != "" {
		return map[string]any{"unresolvedRef": schema.Ref}
	}
	if schema.Example != nil {
		return boundedExample(schema.Example, depth)
	}
	if schema.Default != nil {
		return boundedExample(schema.Default, depth)
	}
	if len(schema.Enum) > 0 {
		return boundedExample(schema.Enum[0], depth)
	}
	switch schema.Type {
	case "boolean":
		return false
	case "integer", "number":
		return 0
	case "array":
		if schema.Items == nil {
			return []any{}
		}
		return []any{schemaExample(*schema.Items, name, depth+1)}
	case "object":
		// fall through to properties below
	}
	if len(schema.Properties) > 0 || schema.Type == "object" {
		keys := make([]string, 0, len(schema.Properties))
		for key := range schema.Properties {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		value := make(map[string]any)
		for i, key := range keys {
			if i >= 20 {
				value["__truncated__"] = "maximum 20 example properties"
				break
			}
			value[key] = schemaExample(schema.Properties[key], key, depth+1)
		}
		return value
	}
	switch schema.Format {
	case "date":
		return "YYYY-MM-DD"
	case "date-time":
		return "YYYY-MM-DDTHH:mm:ssZ"
	case "uuid":
		return "00000000-0000-0000-0000-000000000000"
	}
	if name == "" {
		name = schema.Type
	}
	if name == "" {
		name = "value"
	}
	return "<" + name + ">"
}
func chooseMedia(content map[string]rawMediaType) (string, rawMediaType) {
	if len(content) == 0 {
		return "", rawMediaType{}
	}
	if item, ok := content["application/json"]; ok {
		return "application/json", item
	}
	keys := make([]string, 0, len(content))
	for key := range content {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys[0], content[keys[0]]
}
func summarizeRequest(raw rawRequestBody) *RequestSample {
	if raw.Ref == "" && len(raw.Content) == 0 {
		return nil
	}
	output := &RequestSample{Required: raw.Required, UnresolvedRef: raw.Ref}
	if raw.Ref != "" {
		return output
	}
	typ, media := chooseMedia(raw.Content)
	output.ContentType = typ
	if media.Example != nil {
		output.Example = boundedExample(media.Example, 0)
	}
	if output.Example == nil {
		output.Example = schemaExample(media.Schema, "body", 0)
	}
	return output
}
func summarizeResponses(raw map[string]rawResponse) []ResponseSummary {
	if len(raw) == 0 {
		return nil
	}
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > 30 {
		keys = keys[:30]
	}
	out := make([]ResponseSummary, 0, len(keys))
	for _, code := range keys {
		item := raw[code]
		value := ResponseSummary{
			Status: code, Description: item.Description, UnresolvedRef: item.Ref,
		}
		if len(value.Description) > 400 {
			value.Description = value.Description[:400]
		}
		if item.Ref == "" && len(item.Content) > 0 {
			typ, media := chooseMedia(item.Content)
			value.ContentType = typ
			if media.Example != nil {
				value.Example = boundedExample(media.Example, 0)
			}
			if value.Example == nil {
				value.Example = schemaExample(media.Schema, "response", 0)
			}
		}
		out = append(out, value)
	}
	return out
}
