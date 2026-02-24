package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/bspippi1337/restless/internal/core/app"
)

// Module provides session vars + templating hooks.
type Module struct {
	vars map[string]string
}

func New() *Module {
	return &Module{vars: map[string]string{}}
}

func (m *Module) Name() string { return "session" }

func (m *Module) Register(r *app.Registry) error {
	// Request templating: replace {{var}} in URL and body
	r.RequestMutators = append(r.RequestMutators, func(rc *app.RequestContext) error {
		rc.URL = applyTemplates(rc.URL, m.vars)
		if len(rc.Body) > 0 {
			rc.Body = []byte(applyTemplates(string(rc.Body), m.vars))
		}
		// headers
		for k, vv := range rc.Header {
			for i := range vv {
				vv[i] = applyTemplates(vv[i], m.vars)
			}
			rc.Header[k] = vv
		}
		return nil
	})
	return nil
}

// Set sets a session var (string).
func (m *Module) Set(key, value string) {
	if key == "" {
		return
	}
	m.vars[key] = value
}

// ExtractJSON extracts a value from JSON response by a simple dot path (v1).
func (m *Module) ExtractJSON(dotPath string, body []byte) (string, error) {
	if dotPath == "" {
		return "", errors.New("empty path")
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return "", err
	}
	parts := bytes.Split([]byte(dotPath), []byte("."))
	cur := v
	for _, p := range parts {
		key := string(p)
		obj, ok := cur.(map[string]any)
		if !ok {
			return "", errors.New("path not found")
		}
		cur, ok = obj[key]
		if !ok {
			return "", errors.New("path not found")
		}
	}
	switch t := cur.(type) {
	case string:
		return t, nil
	default:
		b, _ := json.Marshal(t)
		return string(b), nil
	}
}

// ExtractRegex extracts first capture group from body text.
func (m *Module) ExtractRegex(pattern string, body []byte) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	mm := re.FindSubmatch(body)
	if len(mm) < 2 {
		return "", errors.New("no match")
	}
	return string(mm[1]), nil
}

var tmplRe = regexp.MustCompile(`\{\{([a-zA-Z0-9_.-]+)\}\}`)

func applyTemplates(s string, vars map[string]string) string {
	return tmplRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := tmplRe.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		if v, ok := vars[sub[1]]; ok {
			return v
		}
		return m
	})
}
