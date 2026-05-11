package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/recon"
	"github.com/bspippi1337/restless/internal/topology"
	"github.com/spf13/cobra"
)

type BlckswanReport struct {
	Target    string             `json:"target"`
	Generated string             `json:"generated_at"`
	Seeds     []string           `json:"seeds"`
	Found     []string           `json:"found_paths"`
	OpenAPI   []string           `json:"openapi_paths,omitempty"`
	GraphQL   *recon.GraphQLInfo `json:"graphql,omitempty"`
	Notes     []string           `json:"notes,omitempty"`
	RateHints map[string]string  `json:"rate_hints,omitempty"`
	Topology  string             `json:"topology_ascii"`
}

func NewBlckswanCmd() *cobra.Command {
	var outDir string
	var max int
	var timeout time.Duration
	var wordlist string
	var header []string
	var noGraphQL bool

	cmd := &cobra.Command{
		Use:   "blckswan <url>",
		Short: "Killer recon: auto-discover + OpenAPI + GraphQL + topology + README-ready exports",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, u, err := recon.NormalizeTarget(args[0])
			if err != nil {
				return err
			}

			e := recon.New()
			if timeout > 0 {
				e.Timeout = timeout
			}
			for _, h := range header {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) != 2 {
					continue
				}
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				if k != "" && v != "" {
					e.Headers[k] = v
				}
			}

			seeds := append([]string{}, recon.CommonSeeds...)
			if wordlist != "" {
				b, err := os.ReadFile(wordlist)
				if err != nil {
					return err
				}
				for _, line := range strings.Split(string(b), "\n") {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					if !strings.HasPrefix(line, "/") {
						line = "/" + line
					}
					seeds = append(seeds, line)
				}
			}

			ctx := context.Background()
			found := map[string]bool{}
			openapiPaths := map[string]bool{}
			var rateHints map[string]string
			var notes []string

			// Root probe for same-host URL harvesting
			if rootResp, err := e.Request(ctx, "GET", target, nil); err == nil {
				if len(rootResp.Headers) > 0 {
					rateHints = rootResp.Headers
				}
				if recon.LooksJSON(rootResp.ContentType, rootResp.Body) {
					for _, p := range recon.ExtractSameHostPaths(u.Host, rootResp.Body) {
						found[p] = true
					}
				}
			}

			n := 0
			for _, p := range seeds {
				if max > 0 && n >= max {
					break
				}
				uu := *u
				uu.Path = p
				resp, err := e.Request(ctx, "GET", uu.String(), nil)
				n++
				if err != nil {
					continue
				}
				if resp.Status < 500 {
					found[p] = true
					if recon.LooksJSON(resp.ContentType, resp.Body) {
						paths := recon.TryExtractOpenAPIPaths(resp.Body)
						if len(paths) > 0 {
							notes = append(notes, "openapi-detected:"+p)
							for _, op := range paths {
								openapiPaths[op] = true
								found[op] = true
							}
						}
					}
				}
			}

			var gql *recon.GraphQLInfo
			if !noGraphQL {
				g, _ := recon.TryGraphQLIntrospection(ctx, e, u)
				if g != nil && (g.Introspection || g.Note != "not-found") {
					gql = g
					if g.Introspection {
						found["/graphql"] = true
					}
				}
			}

			fp := make([]string, 0, len(found))
			for p := range found {
				fp = append(fp, p)
			}
			sort.Strings(fp)

			op := make([]string, 0, len(openapiPaths))
			for p := range openapiPaths {
				op = append(op, p)
			}
			sort.Strings(op)

			hostLabel := u.Host
			ascii := topology.ASCII(hostLabel, fp)
			svg := topology.SVG(hostLabel, fp)

			rep := BlckswanReport{
				Target:    target,
				Generated: time.Now().UTC().Format(time.RFC3339),
				Seeds:     seeds,
				Found:     fp,
				OpenAPI:   op,
				GraphQL:   gql,
				Notes:     notes,
				RateHints: rateHints,
				Topology:  ascii,
			}

			if outDir == "" {
				outDir = "dist"
			}
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return err
			}

			stamp := time.Now().UTC().Format("20060102_150405")
			base := fmt.Sprintf("blckswan_%s_%s", u.Host, stamp)

			jsonPath := filepath.Join(outDir, base+".json")
			asciiPath := filepath.Join(outDir, base+".topology.txt")
			svgPath := filepath.Join(outDir, base+".map.svg")
			mdPath := filepath.Join(outDir, base+".summary.md")

			b, _ := json.MarshalIndent(rep, "", "  ")
			_ = os.WriteFile(jsonPath, b, 0o644)
			_ = os.WriteFile(asciiPath, []byte(ascii+"\n"), 0o644)
			_ = os.WriteFile(svgPath, []byte(svg), 0o644)
			_ = os.WriteFile(mdPath, []byte(buildSummaryMarkdown(rep, filepath.Base(svgPath))), 0o644)

			fmt.Println("target:", target)
			if gql != nil {
				fmt.Println("graphql:", gql.Note, "introspection:", gql.Introspection, "types:", gql.Types)
			}
			fmt.Println()
			fmt.Print(ascii)
			fmt.Println()
			fmt.Println("report:", jsonPath)
			fmt.Println("topology:", asciiPath)
			fmt.Println("map:", svgPath)
			fmt.Println("summary:", mdPath)

			return nil
		},
	}

	cmd.Flags().StringVar(&outDir, "out", "dist", "output directory")
	cmd.Flags().IntVar(&max, "max", 120, "max seed probes")
	cmd.Flags().DurationVar(&timeout, "timeout", 6*time.Second, "request timeout")
	cmd.Flags().StringVar(&wordlist, "wordlist", "", "extra paths (one per line)")
	cmd.Flags().StringArrayVar(&header, "header", nil, "extra header (repeatable), e.g. --header 'Authorization: Bearer ...'")
	cmd.Flags().BoolVar(&noGraphQL, "no-graphql", false, "skip GraphQL introspection probe")

	return cmd
}

func buildSummaryMarkdown(rep BlckswanReport, svgFile string) string {
	var b strings.Builder
	b.WriteString("# ⚡ BLCKSWAN Recon Summary\n\n")
	b.WriteString("Target: `" + rep.Target + "`\n\n")
	b.WriteString("Generated: `" + rep.Generated + "`\n\n")

	b.WriteString("## Map\n\n")
	b.WriteString("```html\n")
	b.WriteString("<p align=\"center\"><img src=\"" + svgFile + "\" alt=\"BLCKSWAN API map\"></p>\n")
	b.WriteString("```\n\n")

	b.WriteString("## Topology\n\n")
	b.WriteString("```\n")
	b.WriteString(rep.Topology)
	b.WriteString("```\n\n")

	if len(rep.OpenAPI) > 0 {
		b.WriteString("## OpenAPI extracted paths\n\n")
		for _, p := range rep.OpenAPI {
			b.WriteString("- " + p + "\n")
		}
		b.WriteString("\n")
	}

	if rep.GraphQL != nil {
		b.WriteString("## GraphQL\n\n")
		b.WriteString("- endpoint: `/graphql`\n")
		b.WriteString("- introspection: ")
		if rep.GraphQL.Introspection {
			b.WriteString("yes\n")
		} else {
			b.WriteString("no\n")
		}
		b.WriteString("- note: `" + rep.GraphQL.Note + "`\n\n")
	}

	b.WriteString("## Next moves\n\n")
	b.WriteString("- Rerun with auth header if needed.\n")
	b.WriteString("- Provide a wordlist to expand discovery.\n")
	b.WriteString("- Commit the `.map.svg` for README candy.\n")

	return b.String()
}
