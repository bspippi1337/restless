package fuzzwow

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Probe struct {
	Path   string
	Status int
	Class  string
}

type Result struct {
	Target  string
	Live    []Probe
	Blocked []Probe
	Signals []string
}

var candidates = []string{
	"/graphql",
}

func Fuzz(target string) (*Result, error) {
	if !strings.HasPrefix(target, "http") {
		target = "https://" + target
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	r := &Result{
		Target: target,
	}

	for _, path := range candidates {
		req, _ := http.NewRequest(
			"GET",
			target+path,
			nil,
		)

		req.Header.Set(
			"User-Agent",
			"Restless/420",
		)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		probe := Probe{
			Path:   path,
			Status: resp.StatusCode,
			Class:  classify(resp.StatusCode),
		}

		resp.Body.Close()

		if resp.StatusCode == 200 {
			r.Live = append(r.Live, probe)
		}

		if resp.StatusCode == 401 ||
			resp.StatusCode == 403 {
			r.Blocked = append(r.Blocked, probe)
		}
	}

	if len(r.Blocked) > 0 {
		r.Signals = append(
			r.Signals,
			"restricted administrative surface",
		)

		r.Signals = append(
			r.Signals,
			"github edge shielding",
		)

		r.Signals = append(
			r.Signals,
			"authenticated traversal preferred",
		)
	}

	return r, nil
}

func Render(r *Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\nFUZZ\n")
	fmt.Fprintf(&b, "Target  %s\n\n", trimProto(r.Target))

	if len(r.Signals) > 0 {
		fmt.Fprintf(&b, "Signals\n")
		fmt.Fprintf(&b, "-------\n")

		for _, s := range r.Signals {
			fmt.Fprintf(
				&b,
				"  - %s\n",
				s,
			)
		}

		fmt.Fprintln(&b)
	}

	if len(r.Blocked) > 0 {
		fmt.Fprintf(&b, "Restricted Surface\n")
		fmt.Fprintf(&b, "------------------\n")

		for _, p := range r.Blocked {
			fmt.Fprintf(
				&b,
				"  %-18s %-3d %s\n",
				p.Path,
				p.Status,
				p.Class,
			)
		}
	}

	return b.String()
}

func classify(code int) string {
	switch {
	case code == 200:
		return "live"
	case code == 401:
		return "auth required"
	case code == 403:
		return "restricted"
	default:
		return "unknown"
	}
}

func trimProto(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	return s
}
