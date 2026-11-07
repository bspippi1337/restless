package botanical

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

const catalogVersion = 1
const maxCatalogDocuments = 1000

type Catalog struct {
	Version     int         `json:"version"`
	GeneratedAt time.Time   `json:"generated_at"`
	Documents   []Document  `json:"documents"`
	Operations  []Operation `json:"operations"`
}
type Document struct {
	URL         string   `json:"url"`
	SHA256      string   `json:"sha256"`
	Title       string   `json:"title"`
	Kind        string   `json:"kind"`
	ContentType string   `json:"content_type"`
	Excerpt     string   `json:"excerpt,omitempty"`
	Terms       []string `json:"terms,omitempty"`
}
type Parameter struct {
	Name        string `json:"name" yaml:"name"`
	In          string `json:"in" yaml:"in"`
	Required    bool   `json:"required" yaml:"required"`
	Description string `json:"description,omitempty" yaml:"description"`
	Ref         string `json:"-" yaml:"$ref"`
	Schema      struct {
		Type string `json:"type,omitempty" yaml:"type"`
	} `json:"schema,omitempty" yaml:"schema"`
}
type Operation struct {
	OperationID    string                `json:"operation_id"`
	Method         string                `json:"method"`
	Path           string                `json:"path"`
	Summary        string                `json:"summary,omitempty"`
	Description    string                `json:"description,omitempty"`
	Tags           []string              `json:"tags,omitempty"`
	Parameters     []Parameter           `json:"parameters,omitempty"`
	Security       []map[string][]string `json:"security,omitempty"`
	AuthSpecified  bool                  `json:"auth_specified"`
	UnresolvedRefs bool                  `json:"unresolved_refs"`
	SourceURL      string                `json:"source_url"`
	SourceSHA256   string                `json:"source_sha256"`
	Request *RequestSample `json:"request,omitempty"`
	Responses []ResponseSummary `json:"responses,omitempty"`
}
type rawOperation struct {
	OperationID string                `yaml:"operationId"`
	Summary     string                `yaml:"summary"`
	Description string                `yaml:"description"`
	Tags        []string              `yaml:"tags"`
	Parameters  []Parameter           `yaml:"parameters"`
	Security    []map[string][]string `yaml:"security"`
	RequestBody rawRequestBody `yaml:"requestBody"`
	Responses map[string]rawResponse `yaml:"responses"`
}
type rawPath struct {
	Ref        string        `yaml:"$ref"`
	Parameters []Parameter   `yaml:"parameters"`
	Get        *rawOperation `yaml:"get"`
	Head       *rawOperation `yaml:"head"`
	Options    *rawOperation `yaml:"options"`
	Post       *rawOperation `yaml:"post"`
	Put        *rawOperation `yaml:"put"`
	Patch      *rawOperation `yaml:"patch"`
	Delete     *rawOperation `yaml:"delete"`
	Trace      *rawOperation `yaml:"trace"`
}
type rawSpec struct {
	OpenAPI string `yaml:"openapi"`
	Info    struct {
		Title string `yaml:"title"`
	} `yaml:"info"`
	Security   []map[string][]string `yaml:"security"`
	Paths      map[string]rawPath    `yaml:"paths"`
	Components struct {
		Parameters map[string]Parameter `yaml:"parameters"`
	} `yaml:"components"`
}

var htmlTags = regexp.MustCompile(`(?s)<[^>]*>`)

func BuildCatalog(directory string) (Catalog, error) {
	var catalog Catalog
	catalog.Version = catalogVersion
	catalog.GeneratedAt = time.Now().UTC()
	file, err := os.Open(filepath.Join(directory, "index.jsonl"))
	if err != nil {
		return catalog, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	entries := make(map[string]Record)
	for scanner.Scan() {
		var r Record
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			return catalog, fmt.Errorf("invalid archive manifest: %w", err)
		}
		if r.URL == "" || len(r.SHA256) != 64 {
			return catalog, errors.New("archive manifest has missing provenance")
		}
		if _, err := hex.DecodeString(r.SHA256); err != nil {
			return catalog, fmt.Errorf("invalid document SHA256: %w", err)
		}
		if r.ArchivePath != filepath.Join("sha256", r.SHA256) {
			return catalog, fmt.Errorf("invalid archive path for %s", r.URL)
		}
		entries[r.URL] = r // Latest manifest entry wins without executing or re-fetching anything.
		if len(entries) > maxCatalogDocuments {
			return catalog, fmt.Errorf("catalog document limit exceeded (%d)", maxCatalogDocuments)
		}
	}
	if err := scanner.Err(); err != nil {
		return catalog, err
	}
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		r := entries[key]
		b, err := os.ReadFile(filepath.Join(directory, "sha256", r.SHA256))
		if err != nil {
			return catalog, fmt.Errorf("read %s: %w", r.URL, err)
		}
		if len(b) > int(maxDocumentBytes) {
			return catalog, fmt.Errorf("oversized archived document: %s", r.URL)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != r.SHA256 {
			return catalog, fmt.Errorf("SHA256 mismatch (tampered archive): %s", r.URL)
		}
		d := Document{URL: r.URL, SHA256: r.SHA256, ContentType: r.ContentType, Kind: "reference", Title: r.URL}
		if r.ContentType == "application/pdf" {
			d.Kind = "pdf-unparsed" // Do not hallucinate endpoints from binary documents.
		} else {
			var spec rawSpec
			if err := yaml.Unmarshal(b, &spec); err == nil && strings.HasPrefix(spec.OpenAPI, "3.") {
				d.Kind = "openapi-3"
				if spec.Info.Title != "" {
					d.Title = spec.Info.Title
				}
				for path, p := range spec.Paths {
					if !strings.HasPrefix(path, "/") || p.Ref != "" {
						continue
					} // Local-only references are not resolved.
					choices := []struct {
						method string
						op     *rawOperation
					}{
						{"GET", p.Get}, {"HEAD", p.Head}, {"OPTIONS", p.Options}, {"POST", p.Post},
						{"PUT", p.Put}, {"PATCH", p.Patch}, {"DELETE", p.Delete}, {"TRACE", p.Trace},
					}
					for _, choice := range choices {
						if choice.op == nil {
							continue
						}
						op := choice.op
						pathParams, pathUnresolved := resolveParameters(p.Parameters, spec.Components.Parameters)
						opParams, opUnresolved := resolveParameters(op.Parameters, spec.Components.Parameters)
						parameters := mergeParameters(pathParams, opParams)
						security := op.Security
						if security == nil {
							security = spec.Security
						}
						unresolved := op.RequestBody.Ref != "" || pathUnresolved || opUnresolved
						for _, param := range parameters {
							if param.Ref != "" {
								unresolved = true
							}
						}
						catalog.Operations = append(catalog.Operations, Operation{
							OperationID: op.OperationID, Method: choice.method, Path: path, Summary: op.Summary,
							Description: op.Description, Tags: op.Tags, Parameters: parameters, Security: security,
							AuthSpecified: len(security) > 0, UnresolvedRefs: unresolved,
							SourceURL: r.URL, SourceSHA256: r.SHA256,
							Request: summarizeRequest(op.RequestBody), Responses: summarizeResponses(op.Responses),
						})
					}
				}
			} else {
				text := string(b)
				if r.ContentType == "text/html" {
					text = html.UnescapeString(htmlTags.ReplaceAllString(text, " "))
				}
				text = strings.Join(strings.Fields(text), " ")
				if len(text) > 4096 {
					text = text[:4096]
				}
				d.Excerpt = text
				d.Terms = tokens(text)
			}
		}
		catalog.Documents = append(catalog.Documents, d)
	}
	sort.Slice(catalog.Operations, func(i, j int) bool {
		a, b := catalog.Operations[i], catalog.Operations[j]
		if a.SourceURL != b.SourceURL {
			return a.SourceURL < b.SourceURL
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Method < b.Method
	})
	return catalog, nil
}

func mergeParameters(path, operation []Parameter) []Parameter {
	merged := append([]Parameter(nil), path...)
	for _, p := range operation {
		found := false
		for i, existing := range merged {
			if existing.Name == p.Name && existing.In == p.In && p.Ref == "" {
				merged[i] = p
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, p)
		}
	}
	return merged
}

func SaveCatalog(dir string, c Catalog) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	target := filepath.Join(dir, "catalog.json")
	f, err := os.CreateTemp(dir, ".catalog-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), target)
}
func LoadCatalog(dir string) (Catalog, error) {
	var c Catalog
	f, err := os.Open(filepath.Join(dir, "catalog.json"))
	if err != nil {
		return c, err
	}
	defer f.Close()
	if err = json.NewDecoder(f).Decode(&c); err != nil {
		return c, err
	}
	if c.Version != catalogVersion {
		return c, fmt.Errorf("unsupported catalog version %d", c.Version)
	}
	return c, nil
}

func tokens(s string) []string {
	var words []string
	var current []rune
	flush := func() {
		if len(current) == 0 {
			return
		}
		w := strings.ToLower(string(current))
		if len(w) > 4 && strings.HasSuffix(w, "s") {
			w = strings.TrimSuffix(w, "s")
		}
		if len(w) > 1 {
			words = append(words, w)
		}
		current = nil
	}
	var prev rune
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			flush()
			prev = 0
			continue
		}
		if len(current) > 0 && unicode.IsUpper(r) && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
			flush()
		}
		current = append(current, r)
		prev = r
	}
	flush()
	seen := map[string]bool{}
	var out []string
	for _, w := range words {
		if !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	sort.Strings(out)
	return out
}
