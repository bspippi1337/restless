package restlesscore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Endpoint struct {
	Method string
	Path string
	Status int
	Confidence string
	Source string
}

type Edge struct {
	From string
	To string
}

type ScanResult struct {
	Target string
	BaseURL string
	APIType string
	Fingerprints []string
	Confirmed []Endpoint
	Topology []Edge
}

type CrawlNode struct {
	Path string
	Depth int
}

func Render(title string, r *ScanResult) string {
	var b strings.Builder

	live := 0
	gated := 0

	for _, ep := range r.Confirmed {
		if ep.Status >= 200 && ep.Status < 300 {
			live++
		} else {
			gated++
		}
	}

	line := strings.Repeat("═", 60)
	section := strings.Repeat("─", 60)

	fmt.Fprintf(&b, "RESTLESS ENGINE\n")
	fmt.Fprintf(&b, "%s\n\n", line)
	fmt.Fprintf(&b, "TARGET      %s\n", trimProto(r.Target))
	fmt.Fprintf(&b, "TYPE        %s\n", r.APIType)
	fmt.Fprintf(&b, "TRAITS      %s\n\n", strings.Join(r.Fingerprints, " · "))

	fmt.Fprintf(&b, "SURFACE\n")
	fmt.Fprintf(&b, "%s\n\n", section)

	groups := map[string][]Endpoint{}

	for _, ep := range r.Confirmed {
		if ep.Path == "/" {
			continue
		}

		groups[groupName(ep.Path)] = append(groups[groupName(ep.Path)], ep)
	}

	order := []string{"IDENTITY", "REPOSITORIES", "ACTIVITY", "SEARCH", "PLATFORM", "MISC"}

	for _, name := range order {
		items := groups[strings.Title(strings.ToLower(name))]
		if len(items) == 0 {
			continue
		}

		fmt.Fprintf(&b, "%s\n", name)

		for _, ep := range items {
			state := "gated"
			if ep.Status >= 200 && ep.Status < 300 {
				state = "live"
			}

			fmt.Fprintf(&b, "  %-30s %s\n", dotted(strings.TrimPrefix(ep.Path, "/"), 30), state)
		}

		fmt.Fprintf(&b, "\n")
	}

	fmt.Fprintf(&b, "SIGNALS\n")
	fmt.Fprintf(&b, "%s\n\n", section)
	fmt.Fprintf(&b, "  authenticated escalation preferred\n")
	fmt.Fprintf(&b, "  public edge heavily shielded\n")
	fmt.Fprintf(&b, "  anonymous traversal partially available\n\n")

	fmt.Fprintf(&b, "CAPABILITY\n")
	fmt.Fprintf(&b, "%s\n\n", section)
	fmt.Fprintf(&b, "  %-30s %d\n", dotted("discovered", 30), len(r.Confirmed))
	fmt.Fprintf(&b, "  %-30s %d\n", dotted("live", 30), live)
	fmt.Fprintf(&b, "  %-30s %d\n", dotted("gated", 30), gated)
	fmt.Fprintf(&b, "  %-30s medium\n\n", dotted("attack surface", 30))

	fmt.Fprintf(&b, "WORKFLOWS\n")
	fmt.Fprintf(&b, "%s\n\n", section)
	fmt.Fprintf(&b, "  restless discover %s\n", trimProto(r.Target))
	fmt.Fprintf(&b, "  restless inspect  %s\n", trimProto(r.Target))
	fmt.Fprintf(&b, "  restless fuzz     %s\n", trimProto(r.Target))
	fmt.Fprintf(&b, "  restless map      %s\n", trimProto(r.Target))

	return b.String()
}

func dotted(s string, width int) string {
	if len(s) >= width {
		return s
	}

	return s + strings.Repeat(".", width-len(s))
}

func trimProto(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	return strings.TrimRight(s, "/")
}
