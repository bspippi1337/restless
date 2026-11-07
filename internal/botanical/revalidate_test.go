package botanical

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConditionalETagAvoidsFullRedownload(t *testing.T) {
	const document = `{"openapi":"3.0.4","info":{"title":"Sample"},"paths":{"/pets":{"get":{"operationId":"listPets"}}}}`
	const lastModified = "Wed, 21 Oct 2015 07:28:00 GMT"
	fullResponses, unchangedResponses := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/robots.txt":
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
		case "/docs/openapi.json":
			if r.Header.Get("If-None-Match") == `"rev1"` && r.Header.Get("If-Modified-Since") == lastModified {
				unchangedResponses++
				w.WriteHeader(http.StatusNotModified)
				return
			}
			fullResponses++
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", `"rev1"`)
			w.Header().Set("Last-Modified", lastModified)
			_, _ = w.Write([]byte(document))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	config := Config{
		SeedURL:     server.URL + "/docs/openapi.json",
		AllowedHost: parsed.Hostname(),
		Agent:       "RestlessBot/42",
		OutputDir:   dir,
		MaxPages:    1,
		MaxDepth:    0,
		Authorized:  true,
		Client:      server.Client(),
	}
	first, err := Crawl(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || first[0].ResponseStatus != http.StatusOK || first[0].ETag == "" {
		t.Fatalf("first crawl: %+v", first)
	}
	second, err := Crawl(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 1 || second[0].ResponseStatus != http.StatusNotModified {
		t.Fatalf("second crawl: %+v", second)
	}
	if fullResponses != 1 || unchangedResponses != 1 {
		t.Fatalf("expected one full document and one 304, got full=%d, 304=%d", fullResponses, unchangedResponses)
	}
	if first[0].SHA256 != second[0].SHA256 {
		t.Fatal("revalidated document changed archive digest")
	}
	catalog, err := BuildCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Operations) != 1 || catalog.Operations[0].OperationID != "listPets" {
		t.Fatalf("catalog lost revalidated content: %+v", catalog)
	}
}

func TestTamperedCacheStopsBeforeDocumentRequest(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
		calls++
		w.Header().Set("ETag", `"old"`)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("Original API documentation"))
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	config := Config{
		SeedURL:     server.URL + "/docs/api.txt",
		AllowedHost: u.Hostname(),
		Agent:       "RestlessBot/42",
		OutputDir:   dir,
		MaxPages:    1,
		Authorized:  true,
		Client:      server.Client(),
	}
	first, err := Crawl(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, first[0].ArchivePath), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = Crawl(context.Background(), config)
	if err == nil || !strings.Contains(err.Error(), "corrupt cache") {
		t.Fatalf("expected fail-closed cache rejection, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("unexpected additional document request; calls=%d", calls)
	}
}

func TestUnexpected304FailsClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Crawl(context.Background(), Config{
		SeedURL:     server.URL + "/docs/api.json",
		AllowedHost: u.Hostname(),
		Agent:       "RestlessBot/42",
		OutputDir:   t.TempDir(),
		MaxPages:    1,
		Authorized:  true,
		Client:      server.Client(),
	})
	if err == nil || !strings.Contains(err.Error(), "unexpected HTTP 304") {
		t.Fatalf("expected 304 rejection, got %v", err)
	}
}

func TestBoundedValidators(t *testing.T) {
	if boundedETag("abc\r\nInjected: x") != "" {
		t.Fatal("accepted header injection")
	}
	if boundedETag(strings.Repeat("x", 257)) != "" {
		t.Fatal("accepted oversized ETag")
	}
	if boundedLastModified("yesterday") != "" {
		t.Fatal("accepted invalid HTTP date")
	}
	if boundedLastModified("Wed, 21 Oct 2015 07:28:00 GMT") == "" {
		t.Fatal("rejected valid date")
	}
}
