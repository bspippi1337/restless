package library

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bspippi1337/restless/internal/botanical"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func respond(status int, kind, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{kind}}, Body: io.NopCloser(strings.NewReader(body))}
}
func TestDomainNormalization(t *testing.T) {
	for input, want := range map[string]string{
		"api.example.org":                         "https://api.example.org/",
		"HTTPS://Docs.Example.Org/docs/reference": "https://docs.example.org/docs/reference",
		"docs.example.org/openapi.json":           "https://docs.example.org/openapi.json",
	} {
		got, err := Normalize(input)
		if err != nil {
			t.Fatal(input, err)
		}
		if got.SeedURL != want {
			t.Errorf("%s: got %s want %s", input, got.SeedURL, want)
		}
	}
	for _, input := range []string{
		"http://example.org/", "https://127.0.0.1/", "localhost", "https://internal.local/",
		"https://example.org:443/", "https://user:pass@example.org/",
		"https://example.org/?token=foo", "https://example.org/#hello",
		"https://test.invalid/", "https://[::1]/", "https://a..com/", ".example.org", "example",
	} {
		if _, err := Normalize(input); err == nil {
			t.Errorf("accepted unsafe target %q", input)
		}
	}
}
func TestDryRunAndAuthorization(t *testing.T) {
	called := false
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return respond(200, "text/plain", "unexpected"), nil
	})}
	summary, err := Build(context.Background(), Options{Input: "docs.example.org", MaxPages: 4, MaxDepth: 2, Client: client})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("dry run made network request")
	}
	if summary.Mode != "DRY_RUN" || summary.NetworkRequests != "0" {
		t.Fatalf("%+v", summary)
	}
	if summary.OutputDir != filepath.Join(".restless", "libraries", "docs.example.org") {
		t.Fatal(summary.OutputDir)
	}
	_, err = Build(context.Background(), Options{Input: "docs.example.org", Execute: true, MaxPages: 2, Client: client})
	if err == nil || !strings.Contains(err.Error(), "authorization") {
		t.Fatalf("consent bypass: %v", err)
	}
}
func TestRealArchiveCatalogAndLinkedSources(t *testing.T) {
	calls := []string{}
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, r.URL.Path)
		switch r.URL.Path {
		case "/robots.txt":
			return respond(200, "text/plain", "User-agent: *\nAllow: /\n"), nil
		case "/":
			return respond(200, "text/html", "<a href=\"/docs/openapi.json\">API docs</a><a href=\"https://evil.invalid/docs\">external</a>"), nil
		case "/docs/openapi.json":
			return respond(200, "application/json", "{\"openapi\":\"3.0.3\",\"info\":{\"title\":\"Acme API\"},\"paths\":{\"/users\":{\"get\":{\"operationId\":\"listUsers\"},\"post\":{\"operationId\":\"createUser\"}}}}"), nil
		default:
			return respond(404, "text/plain", "missing"), nil
		}
	})
	events := 0
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	result, err := Build(ctx, Options{
		Input: "docs.example.org", OutputDir: t.TempDir(), MaxPages: 3, MaxDepth: 2,
		Execute: true, Authorized: true, Client: &http.Client{Transport: transport},
		OnDocument: func(_ botanical.Record) { events++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 {
		t.Fatalf("expected robots/root/linked doc, got %v", calls)
	}
	if events != 2 || result.DocumentCount != 2 || result.OperationCount != 2 {
		t.Fatalf("wrong counts: progress %d, docs %d, operations %d", events, result.DocumentCount, result.OperationCount)
	}
	if result.Methods["GET"] != 1 || result.Methods["POST"] != 1 {
		t.Fatalf("wrong methods %v", result.Methods)
	}
	for _, name := range []string{"library.json", "catalog.json", "index.jsonl"} {
		if _, err := os.Stat(filepath.Join(result.OutputDir, name)); err != nil {
			t.Fatal(name, err)
		}
	}
	if len(result.Records) != 2 || result.Records[1].SHA256 == "" {
		t.Fatal("missing provenance")
	}
}
