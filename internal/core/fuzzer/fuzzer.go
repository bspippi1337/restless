package fuzzer

import (
	"regexp"
	"strings"

	"github.com/bspippi1337/restless/internal/core/model"
)

type Options struct {
	MaxExtra int
}

func Expand(seeds []model.Endpoint, opt Options) []model.Endpoint {
	maxExtra := opt.MaxExtra
	if maxExtra <= 0 {
		maxExtra = 50
	}

	seen := map[string]bool{}
	add := func(m, p string, out *[]model.Endpoint) {
		k := strings.ToUpper(m) + " " + p
		if seen[k] {
			return
		}
		seen[k] = true
		*out = append(*out, model.Endpoint{Method: strings.ToUpper(m), Path: p})
	}

	out := []model.Endpoint{}
	for _, s := range seeds {
		p := strings.TrimSpace(s.Path)
		if p == "" || !strings.HasPrefix(p, "/") {
			continue
		}
		seen[strings.ToUpper(s.Method)+" "+p] = true
	}

	reVar := regexp.MustCompile(`\{[^}]+\}`)
	reLeaf := regexp.MustCompile(`/([a-z0-9_\-]+)$`)

	for _, s := range seeds {
		if len(out) >= maxExtra {
			break
		}
		p := strings.TrimSpace(s.Path)
		if p == "" || !strings.HasPrefix(p, "/") {
			continue
		}

		if reVar.MatchString(p) {
			base := reVar.ReplaceAllString(p, "")
			base = strings.TrimSuffix(base, "/")
			if base != "" {
				add("GET", base, &out)
				add("GET", base+"/list", &out)
			}
		}

		if strings.Count(p, "/") >= 1 && !reVar.MatchString(p) {
			low := strings.ToLower(p)
			if strings.Contains(low, "health") || strings.Contains(low, "status") || strings.Contains(low, "version") {
				continue
			}
			add("GET", p+"/{id}", &out)
		}

		if m := reLeaf.FindStringSubmatch(p); len(m) == 2 {
			prefix := strings.TrimSuffix(p, "/"+m[1])
			if strings.HasPrefix(prefix, "/v") && len(out) < maxExtra {
				add("GET", prefix+"/health", &out)
				add("GET", prefix+"/status", &out)
			}
		}
	}

	for _, p := range []string{"/health", "/status", "/version"} {
		if len(out) >= maxExtra {
			break
		}
		add("GET", p, &out)
	}
	return out
}
