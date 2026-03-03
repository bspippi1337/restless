package report

import (
	"encoding/json"
	"fmt"

	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
)

type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version,omitempty"`
	Rules   []sarifRule `json:"rules,omitempty"`
}

type sarifRule struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ShortDescription struct {
		Text string `json:"text"`
	} `json:"shortDescription"`
}

type sarifResult struct {
	RuleID  string `json:"ruleId"`
	Level   string `json:"level"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []sarifLocation `json:"locations,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion          `json:"region,omitempty"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

func ToSARIF(appVersion string, res model.GuardResult) ([]byte, error) {
	rules := []sarifRule{
		rule("contract.extra_field", "Extra field not in contract"),
		rule("contract.missing_field", "Missing required field"),
		rule("contract.type_mismatch", "Type mismatch vs contract"),
		rule("contract.enum_violation", "Enum violation vs contract"),
		rule("contract.schema_violation", "Schema violation vs contract"),
	}

	var results []sarifResult
	for _, f := range res.Findings {
		rid := "contract.schema_violation"
		switch f.Kind {
		case model.KindExtraField:
			rid = "contract.extra_field"
		case model.KindMissingField:
			rid = "contract.missing_field"
		case model.KindTypeMismatch:
			rid = "contract.type_mismatch"
		case model.KindEnumViolation:
			rid = "contract.enum_violation"
		}
		level := mapLevel(f.Severity)
		text := fmt.Sprintf("%s %s %d %s: %s (%s)", f.Method, f.Path, f.Status, f.JSONPath, f.Message, f.OpID)

		var r sarifResult
		r.RuleID = rid
		r.Level = level
		r.Message.Text = text
		r.Locations = []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: res.SpecRef},
				Region:           &sarifRegion{StartLine: 1},
			},
		}}
		results = append(results, r)
	}

	log := sarifLog{
		Version: "2.1.0",
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json",
		Runs: []sarifRun{{
			Tool: sarifTool{Driver: sarifDriver{
				Name:    "restless-openapi-guard",
				Version: appVersion,
				Rules:   rules,
			}},
			Results: results,
		}},
	}
	return json.MarshalIndent(log, "", "  ")
}

func rule(id, desc string) sarifRule {
	var r sarifRule
	r.ID = id
	r.Name = id
	r.ShortDescription.Text = desc
	return r
}

func mapLevel(s model.FindingSeverity) string {
	switch s {
	case model.SevInfo, model.SevLow:
		return "note"
	case model.SevMedium:
		return "warning"
	case model.SevHigh, model.SevCritical:
		return "error"
	default:
		return "warning"
	}
}
