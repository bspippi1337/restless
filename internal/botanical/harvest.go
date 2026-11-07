package botanical

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "html"
    "io"
    "net"
    "net/http"
    "net/netip"
    "net/url"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"
)

const maxDocumentBytes int64 = 2 << 20
const maxPagesHard = 25

type Config struct {
    SeedURL     string
    AllowedHost string
    Agent       string
    OutputDir   string
    MaxPages    int
    MaxDepth    int
    Delay       time.Duration
    Authorized  bool
    Client      *http.Client // Tests may inject a local client. Production defaults to a public-IP-only dialer.
}

type Record struct {
    URL         string    `json:"url"`
    Source      string    `json:"source"`
    Status      int       `json:"status"`
    ContentType string    `json:"content_type"`
    SHA256      string    `json:"sha256"`
    ArchivePath string    `json:"archive_path"`
    Bytes       int       `json:"bytes"`
    Depth       int       `json:"depth"`
    FetchedAt   time.Time `json:"fetched_at"`
}

type queued struct {
    url    *url.URL
    depth  int
    source string
}

var hrefRE = regexp.MustCompile(`(?i)\bhref\s*=\s*["']([^"']+)["']`)

func Crawl(ctx context.Context, cfg Config) ([]Record, error) {
    if !cfg.Authorized {
        return nil, errors.New("network access requires explicit --execute and --authorized flags")
    }
    start, err := url.Parse(cfg.SeedURL)
    if err != nil || start == nil || start.Hostname() == "" {
        return nil, errors.New("invalid seed URL")
    }
    if start.User != nil || start.RawQuery != "" || start.Fragment != "" {
        return nil, errors.New("seed URL must not contain credentials, query parameters or a fragment")
    }
    if start.Scheme != "https" && !(cfg.Client != nil && start.Scheme == "http") {
        return nil, errors.New("HTTPS is required for network crawling")
    }
    if cfg.AllowedHost == "" || !strings.EqualFold(start.Hostname(), cfg.AllowedHost) {
        return nil, errors.New("explicit --host must match seed hostname")
    }
    if cfg.Agent == "" || strings.TrimSpace(cfg.Agent) == "" {
        return nil, errors.New("an identifiable user-agent is required")
    }
    if cfg.MaxPages < 1 || cfg.MaxPages > maxPagesHard {
        return nil, fmt.Errorf("max pages must be between 1 and %d", maxPagesHard)
    }
    if cfg.MaxDepth < 0 || cfg.MaxDepth > 3 {
        return nil, errors.New("max depth must be 0..3")
    }
    if cfg.Delay < 2*time.Second {
        cfg.Delay = 2 * time.Second
    }
    if cfg.OutputDir == "" { cfg.OutputDir = ".restless/botanical" }

    client := cfg.Client
    if client == nil { client = safeClient() }
    origin := start.Scheme + "://" + start.Host
    robotsURL := origin + "/robots.txt"
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
    if err != nil { return nil, err }
    req.Header.Set("User-Agent", cfg.Agent)
    resp, err := client.Do(req)
    if err != nil { return nil, fmt.Errorf("robots.txt unavailable (fail closed): %w", err) }
    content, err := io.ReadAll(io.LimitReader(resp.Body, 500<<10+1))
    resp.Body.Close()
    if err != nil { return nil, err }
    var policy robotsPolicy
    switch resp.StatusCode {
    case http.StatusOK:
        policy, err = parseRobots(content)
        if err != nil { return nil, err }
    case http.StatusNotFound, http.StatusGone:
        // RFC 9309: robots.txt is unavailable; no rules apply.
    default:
        return nil, fmt.Errorf("robots.txt status %d; crawl blocked (fail closed)", resp.StatusCode)
    }
    _, robotDelay := policy.selected(cfg.Agent)
    if robotDelay > cfg.Delay { cfg.Delay = robotDelay }
    last := time.Now()

    start.Fragment = ""
    start.RawQuery = ""
    if start.Path == "" { start.Path = "/" }
    queue := []queued{{start, 0, "seed"}}
    seen := map[string]bool{}
    var records []Record

    for len(queue) > 0 && len(records) < cfg.MaxPages {
        item := queue[0]
        queue = queue[1:]
        key := item.url.String()
        if seen[key] { continue }
        seen[key] = true
        if !sameOrigin(start, item.url) || !policy.allowed(cfg.Agent, item.url.EscapedPath()) { continue }

        if pause := cfg.Delay - time.Since(last); pause > 0 {
            timer := time.NewTimer(pause)
            select {
            case <-ctx.Done():
                timer.Stop()
                return records, ctx.Err()
            case <-timer.C:
            }
        }
        request, err := http.NewRequestWithContext(ctx, http.MethodGet, key, nil)
        if err != nil { continue }
        request.Header.Set("User-Agent", cfg.Agent)
        request.Header.Set("Accept", "text/html, application/json, application/pdf, text/plain, application/yaml, text/yaml")
        request.Header.Set("Cache-Control", "no-cache")
        response, err := client.Do(request)
        last = time.Now()
        if err != nil { return records, fmt.Errorf("request %s: %w", key, err) }
        if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
            response.Body.Close()
            return records, fmt.Errorf("server requested pause (HTTP %d): stopped without retrying", response.StatusCode)
        }
        if response.StatusCode != http.StatusOK {
            response.Body.Close()
            continue
        }
        ct := strings.ToLower(strings.TrimSpace(strings.SplitN(response.Header.Get("Content-Type"), ";", 2)[0]))
        if !supportedType(ct) {
            response.Body.Close()
            continue
        }
        body, err := io.ReadAll(io.LimitReader(response.Body, maxDocumentBytes+1))
        response.Body.Close()
        if err != nil { return records, err }
        if int64(len(body)) > maxDocumentBytes {
            continue // Never archive a partial document.
        }
        digest := sha256.Sum256(body)
        hash := hex.EncodeToString(digest[:])
        archivePath := filepath.Join("sha256", hash)
        if err := os.MkdirAll(filepath.Join(cfg.OutputDir, "sha256"), 0700); err != nil { return records, err }
        if err := os.WriteFile(filepath.Join(cfg.OutputDir, archivePath), body, 0600); err != nil { return records, err }
        record := Record{
            URL: key, Source: item.source, Status: response.StatusCode,
            ContentType: ct, SHA256: hash, ArchivePath: archivePath,
            Bytes: len(body), Depth: item.depth, FetchedAt: time.Now().UTC(),
        }
        if err := appendRecord(filepath.Join(cfg.OutputDir, "index.jsonl"), record); err != nil { return records, err }
        records = append(records, record)
        if ct != "text/html" || item.depth >= cfg.MaxDepth { continue }
        for _, match := range hrefRE.FindAllSubmatch(body, -1) {
            href := strings.TrimSpace(html.UnescapeString(string(match[1])))
            next, err := item.url.Parse(href)
            if err != nil || next.User != nil || next.RawQuery != "" || next.Fragment != "" { continue }
            if !sameOrigin(start, next) || !likelyDocumentation(next.Path) { continue }
            if !seen[next.String()] {
                queue = append(queue, queued{next, item.depth + 1, key})
            }
        }
    }
    return records, nil
}

func appendRecord(path string, record Record) error {
    b, err := json.Marshal(record)
    if err != nil { return err }
    file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil { return err }
    defer file.Close()
    _, err = file.Write(append(b, '\n'))
    return err
}

func supportedType(ct string) bool {
    switch ct {
    case "text/html", "text/plain", "text/markdown", "text/yaml",
        "application/json", "application/pdf", "application/yaml",
        "application/x-yaml", "application/vnd.oai.openapi+json":
        return true
    }
    return false
}
func likelyDocumentation(path string) bool {
    p := strings.ToLower(path)
    for _, bad := range []string{"/logout", "/delete", "/admin", "/account", "/checkout"} {
        if strings.Contains(p, bad) { return false }
    }
    for _, prefix := range []string{"/docs", "/documentation", "/developer", "/reference", "/openapi", "/swagger", "/whitepaper", "/papers", "/spec"} {
        if strings.Contains(p, prefix) { return true }
    }
    for _, ext := range []string{".pdf", ".html", ".htm", ".md", ".txt", ".json", ".yaml", ".yml"} {
        if strings.HasSuffix(p, ext) { return true }
    }
    return false
}
func sameOrigin(a, b *url.URL) bool {
    return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host) && b.User == nil && b.RawQuery == ""
}

func safeClient() *http.Client {
    transport := &http.Transport{
        Proxy: nil, // Avoid implicit proxy-based scope expansion.
        DisableKeepAlives: true,
        MaxIdleConns: 1,
        DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
            hostname, port, err := net.SplitHostPort(address)
            if err != nil { return nil, err }
            ips, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
            if err != nil { return nil, err }
            for _, ip := range ips {
                parsed, ok := netip.AddrFromSlice(ip.IP)
                if !ok || !publicIP(parsed.Unmap()) { continue }
                return (&net.Dialer{Timeout: 8 * time.Second}).DialContext(ctx, network, net.JoinHostPort(parsed.String(), port))
            }
            return nil, errors.New("destination has no permitted public IP addresses")
        },
    }
    return &http.Client{
        Timeout: 12 * time.Second,
        Transport: transport,
        CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
            return http.ErrUseLastResponse // Do not follow redirects or new scopes.
        },
    }
}
func publicIP(ip netip.Addr) bool {
    if !ip.IsValid() || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() { return false }
    for _, block := range []string{
        "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24",
        "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24",
        "240.0.0.0/4", "2001:db8::/32",
    } {
        if netip.MustParsePrefix(block).Contains(ip) { return false }
    }
    return true
}
