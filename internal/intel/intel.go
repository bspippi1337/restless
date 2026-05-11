package intel

import (
	"fmt"
	"sort"
	"strings"
)

type Endpoint struct {
	Path   string
	Status int
	Source string
}

type Pulse struct {
	Name   string
	Weight int
	Reason string
}

type Reflex struct {
	Trigger string
	Action  string
	Reason  string
}

type Profile struct {
	Target       string
	Kind         string
	Traits       []string
	Endpoints    []Endpoint
	Pulses       []Pulse
	Reflexes     []Reflex
	Capabilities []string
	Risks        []string
	Workflows    []string
}

func Analyze(target string, kind string, traits []string, endpoints []Endpoint) Profile {
	p := Profile{
		Target:    normalizeTarget(target),
		Kind:      fallback(kind, "HTTP surface"),
		Traits:    uniq(traits),
		Endpoints: endpoints,
	}

	live := 0
	gated := 0
	search := false
	identity := false
	repo := false
	rate := false
	graphql := false

	for _, ep := range endpoints {
		path := strings.ToLower(ep.Path)

		if ep.Status >= 200 && ep.Status < 300 {
			live++
		} else if ep.Status == 401 || ep.Status == 403 || ep.Status == 405 {
			gated++
		}

		search = search || strings.Contains(path, "search")
		identity = identity || strings.Contains(path, "user")
		repo = repo || strings.Contains(path, "repo")
		rate = rate || strings.Contains(path, "rate")
		graphql = graphql || strings.Contains(path, "graphql")
	}

	if live > 0 {
		p.Pulses = append(p.Pulses, Pulse{"anonymous traversal", 30, "public surface responds without credentials"})
		p.Capabilities = appendUnique(p.Capabilities, "anonymous reconnaissance")
	}

	if gated > 0 {
		p.Pulses = append(p.Pulses, Pulse{"gated expansion", 40, "restricted endpoints detected"})
		p.Reflexes = append(p.Reflexes, Reflex{"403/401 surface", "prefer authenticated traversal", "anonymous view is incomplete"})
		p.Risks = appendUnique(p.Risks, "auth boundary present")
	}

	if search {
		p.Pulses = append(p.Pulses, Pulse{"search cortex", 35, "queryable endpoints detected"})
		p.Capabilities = appendUnique(p.Capabilities, "search enumeration")
		p.Reflexes = append(p.Reflexes, Reflex{"search endpoint", "generate safe query probes", "indexed surfaces reveal structure quickly"})
	}

	if identity {
		p.Capabilities = appendUnique(p.Capabilities, "identity graph traversal")
	}

	if repo {
		p.Capabilities = appendUnique(p.Capabilities, "resource graph traversal")
	}

	if rate {
		p.Pulses = append(p.Pulses, Pulse{"rate governor", 25, "quota endpoint or rate headers observed"})
		p.Reflexes = append(p.Reflexes, Reflex{"rate limit", "slow down and cache", "avoid wasting quota"})
	}

	if graphql {
		p.Pulses = append(p.Pulses, Pulse{"graph nerve", 45, "GraphQL edge detected"})
		p.Risks = appendUnique(p.Risks, "schema exploration surface")
	}

	for _, trait := range p.Traits {
		lt := strings.ToLower(trait)
		if strings.Contains(lt, "github") {
			p.Pulses = append(p.Pulses, Pulse{"github organism", 50, "GitHub API fingerprint detected"})
			p.Reflexes = append(p.Reflexes, Reflex{"github fingerprint", "prefer catalog traversal", "root document exposes route templates"})
		}
		if strings.Contains(lt, "rate") {
			p.Capabilities = appendUnique(p.Capabilities, "quota aware operation")
		}
	}

	p.Workflows = []string{
		"restless discover " + p.Target,
		"restless inspect  " + p.Target,
		"restless fuzz     " + p.Target,
		"restless map      " + p.Target,
	}

	sort.Slice(p.Pulses, func(i, j int) bool {
		return p.Pulses[i].Weight > p.Pulses[j].Weight
	})

	return p
}

func RenderNervousSystem(p Profile) string {
	var b strings.Builder
	bar := strings.Repeat("═", 60)
	thin := strings.Repeat("─", 60)

	fmt.Fprintf(&b, "RESTLESS NERVOUS SYSTEM\n")
	fmt.Fprintf(&b, "%s\n\n", bar)
	fmt.Fprintf(&b, "TARGET      %s\n", p.Target)
	fmt.Fprintf(&b, "ORGANISM    %s\n", p.Kind)
	if len(p.Traits) > 0 {
		fmt.Fprintf(&b, "TRAITS      %s\n", strings.Join(p.Traits, " · "))
	}
	fmt.Fprintf(&b, "\n")

	renderPulses(&b, thin, p.Pulses)
	renderReflexes(&b, thin, p.Reflexes)
	renderList(&b, thin, "CAPABILITIES", p.Capabilities)
	renderList(&b, thin, "RISK SURFACE", p.Risks)
	renderList(&b, thin, "WORKFLOWS", p.Workflows)

	return b.String()
}

func renderPulses(b *strings.Builder, thin string, pulses []Pulse) {
	if len(pulses) == 0 {
		return
	}
	fmt.Fprintf(b, "PULSES\n%s\n\n", thin)
	for _, pulse := range pulses {
		fmt.Fprintf(b, "  %-24s %3d  %s\n", dotted(pulse.Name, 24), pulse.Weight, pulse.Reason)
	}
	fmt.Fprintf(b, "\n")
}

func renderReflexes(b *strings.Builder, thin string, reflexes []Reflex) {
	if len(reflexes) == 0 {
		return
	}
	fmt.Fprintf(b, "REFLEX ARC\n%s\n\n", thin)
	for _, r := range reflexes {
		fmt.Fprintf(b, "  when %-20s -> %s\n", r.Trigger, r.Action)
		fmt.Fprintf(b, "  %-27s %s\n", "", r.Reason)
	}
	fmt.Fprintf(b, "\n")
}

func renderList(b *strings.Builder, thin string, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "%s\n%s\n\n", title, thin)
	for _, item := range uniq(items) {
		fmt.Fprintf(b, "  %s\n", item)
	}
	fmt.Fprintf(b, "\n")
}

func dotted(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(".", width-len(s))
}

func normalizeTarget(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	return strings.TrimRight(s, "/")
}

func fallback(s string, f string) string {
	if strings.TrimSpace(s) == "" {
		return f
	}
	return s
}

func uniq(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func appendUnique(in []string, v string) []string {
	for _, x := range in {
		if x == v {
			return in
		}
	}
	return append(in, v)
}
