package core

import "time"

type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Endpoint struct {
	Method string
	Path   string
}

type VerificationIssue struct {
	Type    string // e.g. missing_required_field, status_mismatch
	Field   string // optional
	Message string
}

type EndpointResult struct {
	Endpoint   Endpoint
	Status     Status
	HTTPStatus int
	Latency    time.Duration
	Issues     []VerificationIssue
}

type Summary struct {
	OK   int
	Warn int
	Fail int
}

type Meta struct {
	RateLimitRemaining int
	RateLimitReset     int64
}

type Insight struct {
	Type    string
	Message string
}

type VerificationResult struct {
	SpecHash string
	BaseURL  string
	Results  []EndpointResult
	Summary  Summary
	Meta     Meta
	Insights []Insight
}
