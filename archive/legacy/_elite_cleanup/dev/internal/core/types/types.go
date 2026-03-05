package types

import "net/http"

// Request is the normalized request model used across CLI/TUI/GUI/modules.
type Request struct {
	Method  string
	URL     string
	Headers http.Header
	Body    []byte
}

// Response is the normalized response model used across CLI/TUI/GUI/modules.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	DurationMs int64
}
