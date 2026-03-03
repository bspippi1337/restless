package openapi

import (
	"os"
	"strings"

	"github.com/bspippi1337/restless/internal/core"
	"gopkg.in/yaml.v3"
)

type Spec struct {
	Paths map[string]map[string]interface{} `yaml:"paths"`
}

func Load(path string) ([]core.Endpoint, error) {
	data, err := loadSource(path)
	if err != nil {
		return nil, err
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	var endpoints []core.Endpoint

	for p, methods := range spec.Paths {
		for m := range methods {
			method := strings.ToUpper(m)

			switch method {
			case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
				endpoints = append(endpoints, core.Endpoint{
					Method: method,
					Path:   p,
				})
			}
		}
	}

	return endpoints, nil
}
