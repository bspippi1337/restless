package model

import "time"

type FindingSeverity string

const (
	SevInfo     FindingSeverity = "info"
	SevLow      FindingSeverity = "low"
	SevMedium   FindingSeverity = "medium"
	SevHigh     FindingSeverity = "high"
	SevCritical FindingSeverity = "critical"
)

type FindingKind string

const (
	KindExtraField      FindingKind = "extra_field"
	KindMissingField    FindingKind = "missing_field"
	KindTypeMismatch    FindingKind = "type_mismatch"
	KindEnumViolation   FindingKind = "enum_violation"
	KindSchemaViolation FindingKind = "schema_violation"
)

type Finding struct {
	OpID        string
	Method      string
	Path        string
	Status      int
	ContentType string

	Kind     FindingKind
	Severity FindingSeverity

	JSONPath string
	Message  string
	Expected string
	Actual   string
}

type GuardResult struct {
	TargetBaseURL string
	SpecRef       string
	StartedAt     time.Time
	FinishedAt    time.Time

	Findings []Finding
	CDI      float64
}

type SemverBump string

const (
	BumpNone  SemverBump = "none"
	BumpPatch SemverBump = "patch"
	BumpMinor SemverBump = "minor"
	BumpMajor SemverBump = "major"
)

type DiffResult struct {
	OldRef string
	NewRef string

	Breaking    []string
	NonBreaking []string

	RecommendedBump SemverBump
}
