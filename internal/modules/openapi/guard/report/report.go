package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
)

func PrintHuman(res model.GuardResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Contract Drift Index: %.3f

", res.CDI)
	for _, f := range res.Findings {
		fmt.Fprintf(&b, "%s %s %d %s [%s/%s]
  %s

",
			f.Method, f.Path, f.Status, f.JSONPath, f.Kind, f.Severity, f.Message)
	}
	if len(res.Findings) == 0 {
		b.WriteString("No contract violations detected.
")
	}
	return b.String()
}

func ToJSON(res model.GuardResult) ([]byte, error) {
	return json.MarshalIndent(res, "", "  ")
}
