package discovery

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/core/docparse"
	"github.com/bspippi1337/restless/internal/core/fuzzer"
	"github.com/bspippi1337/restless/internal/core/model"
	"github.com/bspippi1337/restless/internal/core/probe"
	"github.com/bspippi1337/restless/internal/core/scrape"
)

type Options struct {
	BudgetPages   int
	BudgetSeconds int
	Verify        bool
	Fuzz          bool
}

func DiscoverDomain(domain string, opt Options) (Finding, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return Finding{}, errors.New("domain is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(max(10, opt.BudgetSeconds))*time.Second)
	defer cancel()

	hosts := HostCandidates(domain)
	find := Finding{Hosts: hosts}

	docURLs := []string{}
	endpoints := []Endpoint{}
	base := ""

	for _, h := range hosts {
		oas, urls, err := docparse.TryOpenAPI(ctx, h)
		if err == nil && oas != nil && len(oas.Paths) > 0 {
			if base == "" {
				base = h
			}
			docURLs = append(docURLs, urls...)
			eps := docparse.EndpointsFromOpenAPI(oas)
			for i := range eps {
				eps[i].FullURL = h + eps[i].Path
				eps[i].Evidences = append(eps[i].Evidences, Evidence{
					Source: SourceOpenAPI,
					URL:    urls[0],
					Note:   "OpenAPI (json/yaml)",
					When:   time.Now(),
					Score:  95,
				})
				endpoints = append(endpoints, eps[i])
			}
			break
		}
	}

	if base == "" {
		base = hosts[0]
	}

	smURLs, smPaths := scrape.SitemapDocs(ctx, base, max(12, opt.BudgetPages*3))
	if len(smURLs) > 0 {
		docURLs = append(docURLs, smURLs...)
		for _, p := range smPaths {
			endpoints = append(endpoints, Endpoint{
				Method:  "GET",
				Path:    p,
				FullURL: base + p,
				Evidences: []Evidence{{
					Source: SourceSitemap,
					URL:    base + "/sitemap.xml",
					Note:   "sitemap hint",
					When:   time.Now(),
					Score:  45,
				}},
			})
		}
	}

	scrHits, scrVisited := scrape.LightDocsScrape(ctx, base, max(1, opt.BudgetPages))
	docURLs = append(docURLs, scrVisited...)
	for _, hit := range scrHits {
		url := ""
		if len(scrVisited) > 0 {
			url = scrVisited[0]
		}
		endpoints = append(endpoints, Endpoint{
			Method:  hit.Method,
			Path:    hit.Path,
			FullURL: base + hit.Path,
			Evidences: []Evidence{{
				Source: SourceHTML,
				URL:    url,
				Note:   "docs scrape heuristic",
				When:   time.Now(),
				Score:  55,
			}},
		})
	}

	if opt.Fuzz {

		seed := dedupe(endpoints)

		// Convert to minimal shared shape to avoid import cycles (discovery <-> fuzzer).
		seedModel := make([]model.Endpoint, 0, len(seed))
		for _, s := range seed {
			seedModel = append(seedModel, model.Endpoint{Method: s.Method, Path: s.Path})
		}

		expModel := fuzzer.Expand(seedModel, fuzzer.Options{MaxExtra: 60})
		for _, em := range expModel {
			e := Endpoint{Method: em.Method, Path: em.Path}
			e.Evidences = append(e.Evidences, Evidence{
				Source: SourceFuzzer,
				URL:    base,
				Note:   "seed-only expansion",
				When:   time.Now(),
				Score:  40,
			})
			e.FullURL = base + e.Path
			endpoints = append(endpoints, e)
		}
	}

	endpoints = dedupe(endpoints)

	if opt.Verify {
		verified := []Endpoint{}
		for _, e := range endpoints {
			ok, status, hint := probe.Verify(ctx, e.Method, e.FullURL)
			if ok {
				e.Evidences = append(e.Evidences, Evidence{
					Source: SourceProbe,
					URL:    e.FullURL,
					Note:   "Verified: " + status + " " + hint,
					When:   time.Now(),
					Score:  70,
				})
				verified = append(verified, e)
			}
		}
		if len(verified) > 0 {
			endpoints = verified
		}
	}

	endpoints = dedupe(endpoints)
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path == endpoints[j].Path {
			return endpoints[i].Method < endpoints[j].Method
		}
		return endpoints[i].Path < endpoints[j].Path
	})

	find.BaseURL = base
	find.DocURLs = uniq(docURLs, 24)
	find.Endpoints = endpoints
	find.Notes = append(find.Notes, "Domain-first discovery: docs → scrape → fuzz → verify (safe).")
	return find, nil
}

func HostCandidates(domain string) []string {
	d := strings.TrimSpace(domain)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimSuffix(d, "/")
	d = strings.ToLower(d)

	bases := []string{
		"https://" + d,
		"https://api." + d,
		"https://developer." + d,
		"https://docs." + d,
		"https://sandbox." + d,
		"https://staging." + d,
	}
	return uniq(bases, 12)
}

func dedupe(in []Endpoint) []Endpoint {
	type k struct{ m, p string }
	seen := map[k]Endpoint{}
	for _, e := range in {
		mm := strings.ToUpper(strings.TrimSpace(e.Method))
		pp := strings.TrimSpace(e.Path)
		if mm == "" {
			mm = "GET"
		}
		if pp == "" || !strings.HasPrefix(pp, "/") {
			continue
		}
		kk := k{mm, pp}
		if ex, ok := seen[kk]; ok {
			ex.Evidences = append(ex.Evidences, e.Evidences...)
			seen[kk] = ex
		} else {
			e.Method, e.Path = mm, pp
			seen[kk] = e
		}
	}
	out := make([]Endpoint, 0, len(seen))
	for _, v := range seen {
		out = append(out, v)
	}
	return out
}

func uniq(in []string, maxN int) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
		if maxN > 0 && len(out) >= maxN {
			break
		}
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
