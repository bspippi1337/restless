package fuzzer

import (
	"math/rand"
	"sort"
	"strings"
	"time"
)

type Options struct {
	MaxExtra int
}

// EndpointLite is intentionally tiny to avoid import cycles.
// discovery converts EndpointLite -> discovery.Endpoint.
type EndpointLite struct {
	Method string
	Path   string
}

func ExpandLite(seed []EndpointLite, opt Options) []EndpointLite {
	maxExtra := opt.MaxExtra
	if maxExtra <= 0 {
		maxExtra = 40
	}

	seen := map[string]bool{}
	out := []EndpointLite{}

	add := func(m, p string) {
		m = strings.ToUpper(strings.TrimSpace(m))
		if m == "" {
			m = "GET"
		}
		p = strings.TrimSpace(p)
		if p == "" || !strings.HasPrefix(p, "/") {
			return
		}
		k := m + " " + p
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, EndpointLite{Method: m, Path: p})
	}

	// seed first
	for _, e := range seed {
		add(e.Method, e.Path)
	}

	// build simple vocab from seed paths
	parts := []string{}
	for _, e := range seed {
		for _, seg := range strings.Split(e.Path, "/") {
			seg = strings.TrimSpace(seg)
			if seg == "" {
				continue
			}
			// ignore obvious params
			if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
				continue
			}
			if strings.HasPrefix(seg, ":") {
				continue
			}
			parts = append(parts, seg)
		}
	}
	if len(parts) == 0 {
		parts = []string{"api", "v1", "v2", "health", "status", "users", "auth"}
	}

	// common params candidates
	paramVals := []string{"{id}", "{uuid}", "{userId}", "{projectId}", "{slug}"}

	// candidate methods
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// generate extra endpoints
	for len(out) < len(seen)+maxExtra {
		// random path length 1-4
		n := 1 + rng.Intn(4)
		segs := make([]string, 0, n)
		for i := 0; i < n; i++ {
			if rng.Intn(10) < 2 {
				segs = append(segs, paramVals[rng.Intn(len(paramVals))])
			} else {
				segs = append(segs, parts[rng.Intn(len(parts))])
			}
		}
		p := "/" + strings.Join(segs, "/")
		m := methods[rng.Intn(len(methods))]
		add(m, p)

		// also add GET variant for visibility
		if m != "GET" {
			add("GET", p)
		}

		// stop if we hit limit
		if len(out) >= len(seen)+maxExtra {
			break
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Method < out[j].Method
		}
		return out[i].Path < out[j].Path
	})

	// return only generated beyond original seed (but safe to return all; caller can dedupe)
	return out
}
