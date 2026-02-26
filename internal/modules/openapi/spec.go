package openapi

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Spec struct {
	OpenAPI string `json:"openapi" yaml:"openapi"`
	Info    Info   `json:"info" yaml:"info"`
	Servers []Srv  `json:"servers" yaml:"servers"`
	Paths   Paths  `json:"paths" yaml:"paths"`
}

type Info struct {
	Title   string `json:"title" yaml:"title"`
	Version string `json:"version" yaml:"version"`
}

type Srv struct {
	URL string `json:"url" yaml:"url"`
}

type Paths map[string]PathItem

// PathItem keys are HTTP methods (get/post/put/patch/delete/options/head/trace)
type PathItem map[string]Operation

type Operation struct {
	Summary     string `json:"summary" yaml:"summary"`
	OperationID string `json:"operationId" yaml:"operationId"`
}

func LoadSpecFromFile(path string) (Spec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}
	return LoadSpec(b)
}

func LoadSpec(raw []byte) (Spec, error) {
	// Try JSON first
	var s Spec
	if err := json.Unmarshal(raw, &s); err == nil && s.Paths != nil {
		return validateSpec(s)
	}

	// Then YAML
	if err := yaml.Unmarshal(raw, &s); err == nil && s.Paths != nil {
		return validateSpec(s)
	}

	// Last attempt: YAML may parse into map if weird anchors; still fail fast
	return Spec{}, errors.New("failed to parse spec as JSON or YAML")
}

func validateSpec(s Spec) (Spec, error) {
	if s.Paths == nil {
		return Spec{}, errors.New("invalid spec: missing paths")
	}
	// Normalize method keys to lowercase
	n := Spec{
		OpenAPI: s.OpenAPI,
		Info:    s.Info,
		Servers: s.Servers,
		Paths:   Paths{},
	}
	for p, item := range s.Paths {
		nItem := PathItem{}
		for m, op := range item {
			nItem[strings.ToLower(m)] = op
		}
		n.Paths[p] = nItem
	}
	return n, nil
}

func (s Spec) BaseURL() string {
	if len(s.Servers) == 0 {
		return ""
	}
	return strings.TrimRight(s.Servers[0].URL, "/")
}
