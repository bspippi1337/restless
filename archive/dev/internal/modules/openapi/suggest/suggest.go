package suggest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/modules/openapi/ai"
)

type SuggestionAction string

const (
	ActionMakeOptional SuggestionAction = "make_optional"
	ActionBroadenType  SuggestionAction = "broaden_type"
	ActionExtendEnum   SuggestionAction = "extend_enum"
	ActionAddResponse  SuggestionAction = "add_response"
	ActionReviewSchema SuggestionAction = "review_schema"
)

type Suggestion struct {
	Confidence float64          `json:"confidence"`
	Action     SuggestionAction `json:"action"`
	Selector   Selector         `json:"selector"`
	Note       string           `json:"note"`
	Evidence   Evidence         `json:"evidence"`
}

type Selector struct {
	OperationID string `json:"operation_id"`
	Status      int    `json:"status"`
	Kind        string `json:"kind"`
	JSONPath    string `json:"json_path"`
}

type Evidence struct {
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	AvgCDI    float64   `json:"avg_cdi"`
	Message   string    `json:"message"`
}

type Report struct {
	BaseURL   string       `json:"base_url"`
	SpecRef   string       `json:"spec_ref"`
	Generated time.Time    `json:"generated_at"`
	MinCount  int          `json:"min_count"`
	Items     []Suggestion `json:"items"`
}

func Build(baseURL string, minCount int) (*Report, error) {
	s, err := ai.Load(baseURL)
	if err != nil {
		return nil, fmt.Errorf("no snapshot for %s (run restless against the host first): %w", baseURL, err)
	}

	stats := ai.TopFindings(s, minCount)
	items := make([]Suggestion, 0, len(stats))

	for _, st := range stats {
		action, note := mapKindToSuggestion(st.Key.Kind, st.Key.JSONPath, st.Key.Message)
		conf := confidence(st.Count, st.AvgCDI)

		items = append(items, Suggestion{
			Confidence: conf,
			Action:     action,
			Selector: Selector{
				OperationID: st.Key.OpID,
				Status:      st.Key.Status,
				Kind:        st.Key.Kind,
				JSONPath:    st.Key.JSONPath,
			},
			Note: note,
			Evidence: Evidence{
				Count:     st.Count,
				FirstSeen: st.FirstSeen,
				LastSeen:  st.LastSeen,
				AvgCDI:    st.AvgCDI,
				Message:   st.Key.Message,
			},
		})
	}

	return &Report{
		BaseURL:   s.BaseURL,
		SpecRef:   s.SpecRef,
		Generated: time.Now(),
		MinCount:  minCount,
		Items:     items,
	}, nil
}

func Write(report *Report, outDir string) (mdPath, jsonPath, planPath string, err error) {
	if report == nil {
		return "", "", "", fmt.Errorf("nil report")
	}
	if outDir == "" {
		outDir = "suggestions"
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", "", "", err
	}

	stamp := time.Now().Format("20060102-150405")
	safe := slug(report.BaseURL)
	base := filepath.Join(outDir, fmt.Sprintf("%s-%s", safe, stamp))

	mdPath = base + ".md"
	jsonPath = base + ".json"
	planPath = base + ".plan.json"

	if err := os.WriteFile(mdPath, []byte(ToMarkdown(report)), 0644); err != nil {
		return "", "", "", err
	}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", "", "", err
	}
	if err := os.WriteFile(jsonPath, b, 0644); err != nil {
		return "", "", "", err
	}

	plan := ToPatchPlan(report)
	pb, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return "", "", "", err
	}
	if err := os.WriteFile(planPath, pb, 0644); err != nil {
		return "", "", "", err
	}

	return mdPath, jsonPath, planPath, nil
}

func ToMarkdown(r *Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# OpenAPI Suggestions\n\n")
	fmt.Fprintf(&b, "- Base URL: `%s`\n", r.BaseURL)
	fmt.Fprintf(&b, "- Spec ref: `%s`\n", r.SpecRef)
	fmt.Fprintf(&b, "- Generated: `%s`\n", r.Generated.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Min count: `%d`\n\n", r.MinCount)

	if len(r.Items) == 0 {
		b.WriteString("No repeated drift patterns found above threshold.\n")
		return b.String()
	}

	b.WriteString("## Suggested changes\n\n")

	for i, it := range r.Items {
		fmt.Fprintf(&b, "### %d) %s (confidence %.2f)\n\n", i+1, it.Action, it.Confidence)
		fmt.Fprintf(&b, "- OperationID: `%s`\n", it.Selector.OperationID)
		fmt.Fprintf(&b, "- Status: `%d`\n", it.Selector.Status)
		fmt.Fprintf(&b, "- Kind: `%s`\n", it.Selector.Kind)
		fmt.Fprintf(&b, "- JSONPath: `%s`\n\n", it.Selector.JSONPath)

		fmt.Fprintf(&b, "**Note:** %s\n\n", it.Note)

		fmt.Fprintf(&b, "**Evidence:**\n\n")
		fmt.Fprintf(&b, "- Count: `%d`\n", it.Evidence.Count)
		fmt.Fprintf(&b, "- Avg CDI: `%.3f`\n", it.Evidence.AvgCDI)
		fmt.Fprintf(&b, "- First seen: `%s`\n", it.Evidence.FirstSeen.Format(time.RFC3339))
		fmt.Fprintf(&b, "- Last seen: `%s`\n", it.Evidence.LastSeen.Format(time.RFC3339))
		fmt.Fprintf(&b, "- Message: `%s`\n\n", it.Evidence.Message)
	}

	return b.String()
}

type PatchPlan struct {
	Version string          `json:"version"`
	BaseURL string          `json:"base_url"`
	SpecRef string          `json:"spec_ref"`
	Items   []PatchPlanItem `json:"items"`
}

type PatchPlanItem struct {
	Action     SuggestionAction `json:"action"`
	Confidence float64          `json:"confidence"`
	Selector   Selector         `json:"selector"`
	Note       string           `json:"note"`
}

func ToPatchPlan(r *Report) PatchPlan {
	items := make([]PatchPlanItem, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, PatchPlanItem{
			Action:     it.Action,
			Confidence: it.Confidence,
			Selector:   it.Selector,
			Note:       it.Note,
		})
	}
	return PatchPlan{
		Version: "openapi-suggest-plan/v1",
		BaseURL: r.BaseURL,
		SpecRef: r.SpecRef,
		Items:   items,
	}
}

func mapKindToSuggestion(kind, jsonPath, msg string) (SuggestionAction, string) {
	k := strings.ToLower(kind)
	m := strings.ToLower(msg)

	switch {
	case strings.Contains(k, "missing_field") || strings.Contains(m, "required"):
		return ActionMakeOptional, fmt.Sprintf("Observed missing required field at %s. Consider removing it from required, making it nullable, or adjusting the response schema.", jsonPath)
	case strings.Contains(k, "type_mismatch") || strings.Contains(m, "type"):
		return ActionBroadenType, fmt.Sprintf("Observed type mismatch at %s. Consider changing the field type or using oneOf/anyOf to accept observed variants.", jsonPath)
	case strings.Contains(k, "enum_violation") || strings.Contains(m, "enum"):
		return ActionExtendEnum, fmt.Sprintf("Observed enum drift at %s. Consider extending the enum set to include observed values (or relaxing to string).", jsonPath)
	case strings.Contains(m, "response not defined") || strings.Contains(m, "not defined"):
		return ActionAddResponse, "Observed response code/body not defined in spec. Consider adding the missing response entry."
	default:
		return ActionReviewSchema, "Observed schema drift. Review the response schema for compatibility."
	}
}

func confidence(count int, avgCDI float64) float64 {
	c := float64(count)
	base := 1.0 - (1.0 / (1.0 + c/2.0))
	boost := avgCDI
	if boost > 1.0 {
		boost = 1.0
	}
	out := 0.65*base + 0.35*boost
	if out < 0 {
		return 0
	}
	if out > 1 {
		return 1
	}
	return out
}

func slug(s string) string {
	s = strings.ToLower(s)
	repl := []struct{ old, new string }{
		{"://", "_"},
		{"/", "_"},
		{":", "_"},
		{"?", "_"},
		{"&", "_"},
		{"=", "_"},
	}
	for _, r := range repl {
		s = strings.ReplaceAll(s, r.old, r.new)
	}
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return strings.Trim(s, "_")
}
