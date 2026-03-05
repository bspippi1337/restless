package runtime

import (
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
)

func mapSchemaError(opID, method, path string, status int, contentType string, err error) []model.Finding {
	var out []model.Finding

	if ve, ok := err.(*jsonschema.ValidationError); ok {
		flattenValidationError(opID, method, path, status, contentType, ve, &out)
		return out
	}

	out = append(out, model.Finding{
		OpID:        opID,
		Method:      strings.ToUpper(method),
		Path:        path,
		Status:      status,
		ContentType: contentType,
		Kind:        model.KindSchemaViolation,
		Severity:    model.SevHigh,
		JSONPath:    "$",
		Message:     err.Error(),
	})

	return out
}

func flattenValidationError(
	opID, method, path string,
	status int,
	contentType string,
	ve *jsonschema.ValidationError,
	out *[]model.Finding,
) {
	jp := "$"
	if ve.InstanceLocation != "" {
		jp = "$" + pointerToJSONPath(ve.InstanceLocation)
	}

	msg := ve.Message
	sev := model.SevHigh
	kind := model.KindSchemaViolation

	m := strings.ToLower(msg)

	switch {
	case strings.Contains(m, "required"):
		kind = model.KindMissingField
		sev = model.SevCritical
	case strings.Contains(m, "invalid type"), strings.Contains(m, "type"):
		kind = model.KindTypeMismatch
		sev = model.SevHigh
	case strings.Contains(m, "enum"):
		kind = model.KindEnumViolation
		sev = model.SevMedium
	}

	*out = append(*out, model.Finding{
		OpID:        opID,
		Method:      strings.ToUpper(method),
		Path:        path,
		Status:      status,
		ContentType: contentType,
		Kind:        kind,
		Severity:    sev,
		JSONPath:    jp,
		Message:     msg,
	})

	for _, c := range ve.Causes {
		flattenValidationError(opID, method, path, status, contentType, c, out)
	}
}

func pointerToJSONPath(ptr string) string {
	if ptr == "" || ptr == "/" {
		return ""
	}

	parts := strings.Split(ptr, "/")
	var b strings.Builder

	for _, p := range parts {
		if p == "" {
			continue
		}
		if isDigits(p) {
			b.WriteString("[")
			b.WriteString(p)
			b.WriteString("]")
		} else {
			b.WriteString(".")
			b.WriteString(p)
		}
	}
	return b.String()
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
