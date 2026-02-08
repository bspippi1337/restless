package model

// Endpoint is the minimal shared representation used across modules to avoid import cycles.
// Discovery may enrich endpoints with evidence and full URLs, but the core shape is stable.
type Endpoint struct {
    Method string `json:"method,omitempty"`
    Path   string `json:"path,omitempty"`
}
