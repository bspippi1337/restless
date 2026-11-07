package botanical

import (
    "context"
    "encoding/json"
    "net"
    "net/netip"
    "net/http"
    "net/http/httptest"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestRobotsPolicy(t *testing.T) {
    robots, err := parseRobots([]byte("User-agent: *\nDisallow: /private\nAllow: /private/public\nDisallow: /*.pdf$\nUser-agent: RestlessBot\nDisallow: /secret\nAllow: /secret/docs\n"))
    if err != nil { t.Fatal(err) }
    cases := []struct{ agent, path string; want bool }{
        {"OtherBot", "/private/a", false},
        {"OtherBot", "/private/public", true},
        {"OtherBot", "/whitepaper.pdf", false},
        {"OtherBot", "/whitepaper.pdf?x=1", true},
        {"RestlessBot", "/secret/x", false},
        {"RestlessBot", "/secret/docs", true},
        {"RestlessBot", "/private/a", true}, // specific agent replaces wildcard
    }
    for _, tc := range cases {
        got := robots.allowed(tc.agent, tc.path)
        if got != tc.want { t.Errorf("%s %s: got %v, want %v", tc.agent, tc.path, got, tc.want) }
    }
}

func TestCrawlRequiresAuthorization(t *testing.T) {
    _, err := Crawl(context.Background(), Config{SeedURL: "https://example.org", AllowedHost: "example.org"})
    if err == nil || !strings.Contains(err.Error(), "explicit") { t.Fatalf("got %v", err) }
}

func TestCrawlArchivesOnlyScopedDocuments(t *testing.T) {
    var served []string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        served = append(served, r.URL.Path)
        switch r.URL.Path {
        case "/robots.txt":
            w.Header().Set("Content-Type", "text/plain")
            _, _ = w.Write([]byte("User-agent: *\nDisallow: /docs/hidden\n"))
        case "/":
            w.Header().Set("Content-Type", "text/html")
            _, _ = w.Write([]byte("<a href='/docs/guide.md'>Guide</a><a href='/docs/hidden'>Blocked</a><a href='https://external.example/docs/x'>External</a>"))
        case "/docs/guide.md":
            w.Header().Set("Content-Type", "text/plain")
            _, _ = w.Write([]byte("API reference"))
        default:
            t.Errorf("unexpected request to %s", r.URL.Path)
            http.NotFound(w, r)
        }
    }))
    defer server.Close()
    parsed, err := url.Parse(server.URL)
    if err != nil { t.Fatal(err) }
    out := t.TempDir()
    records, err := Crawl(context.Background(), Config{
        SeedURL: server.URL, AllowedHost: parsed.Hostname(), Authorized: true,
        Agent: "RestlessBot", MaxPages: 3, MaxDepth: 2,
        OutputDir: out, Client: server.Client(),
    })
    if err != nil { t.Fatal(err) }
    if len(records) != 2 { t.Fatalf("got %d records; want 2", len(records)) }
    if len(served) != 3 { t.Fatalf("visited %v", served) }
    data, err := os.ReadFile(filepath.Join(out, "index.jsonl"))
    if err != nil { t.Fatal(err) }
    lines := strings.Split(strings.TrimSpace(string(data)), "\n")
    if len(lines) != 2 { t.Fatalf("got %d index lines", len(lines)) }
    for _, line := range lines {
        var record Record
        if err := json.Unmarshal([]byte(line), &record); err != nil { t.Fatal(err) }
        if _, err := os.Stat(filepath.Join(out, record.ArchivePath)); err != nil { t.Fatal(err) }
    }
}

func TestPublicIPGuard(t *testing.T) {
    blocked := []string{"127.0.0.1", "10.0.0.2", "169.254.169.254", "192.168.1.2", "100.64.0.1", "::1", "2001:db8::1"}
    for _, ip := range blocked {
        parsed := net.ParseIP(ip)
        if parsed == nil { t.Fatalf("bad test address %s", ip) }
        addr, ok := netip.AddrFromSlice(parsed)
        if !ok { t.Fatal(ip) }
        if publicIP(addr.Unmap()) { t.Errorf("should reject %s", ip) }
    }
}
