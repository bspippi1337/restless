package snippets

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Snippet struct {
	Version      int
	Name         string
	Profile      string
	CreatedAt    string
	LastUsedAt   string
	UseCount     int
	SuccessCount int
	FailCount    int
	AvgLatencyMs int64
	Pin          bool
	Tags         []string
	Notes        string

	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

func DefaultDir(profile string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".config", "restless", "snippets", profile)
}

func Save(dir string, s Snippet, overwrite bool) (string, error) {
	if s.Name == "" || s.Profile == "" {
		return "", errors.New("snippet requires name+profile")
	}
	if dir == "" {
		dir = DefaultDir(s.Profile)
	}
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, safeName(s.Name)+".yaml")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return "", errors.New("snippet exists")
		}
	}
	now := time.Now().Format(time.RFC3339)
	if s.Version == 0 {
		s.Version = 1
	}
	if s.CreatedAt == "" {
		s.CreatedAt = now
	}
	if s.LastUsedAt == "" {
		s.LastUsedAt = now
	}

	var b strings.Builder
	b.WriteString("version: 1\n")
	b.WriteString("name: " + s.Name + "\n")
	b.WriteString("profile: " + s.Profile + "\n")
	b.WriteString("createdAt: " + s.CreatedAt + "\n")
	b.WriteString("lastUsedAt: " + s.LastUsedAt + "\n")
	b.WriteString("useCount: " + itoa(s.UseCount) + "\n")
	b.WriteString("successCount: " + itoa(s.SuccessCount) + "\n")
	b.WriteString("failCount: " + itoa(s.FailCount) + "\n")
	b.WriteString("avgLatencyMs: " + itoa64(s.AvgLatencyMs) + "\n")
	b.WriteString("pin: " + boolStr(s.Pin) + "\n")
	if len(s.Tags) > 0 {
		b.WriteString("tags: [" + strings.Join(s.Tags, ", ") + "]\n")
	}
	if strings.TrimSpace(s.Notes) != "" {
		b.WriteString("notes: |\n")
		for _, ln := range strings.Split(s.Notes, "\n") {
			b.WriteString("  " + ln + "\n")
		}
	}
	b.WriteString("\nrequest:\n")
	b.WriteString("  method: " + strings.ToUpper(s.Method) + "\n")
	b.WriteString("  path: " + s.Path + "\n")
	b.WriteString("  headers:\n")
	if len(s.Headers) == 0 {
		b.WriteString("    Accept: application/json\n")
	} else {
		for k, v := range s.Headers {
			b.WriteString("    " + k + ": " + v + "\n")
		}
	}
	if strings.TrimSpace(s.Body) != "" {
		b.WriteString("  body: |\n")
		for _, ln := range strings.Split(s.Body, "\n") {
			b.WriteString("    " + ln + "\n")
		}
	}
	b.WriteString("\n")
	return path, os.WriteFile(path, []byte(b.String()), 0o644)
}

func Load(dir, profile, name string) (Snippet, error) {
	if dir == "" {
		dir = DefaultDir(profile)
	}
	path := filepath.Join(dir, safeName(name)+".yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return Snippet{}, err
	}
	return parse(profile, string(b)), nil
}

func List(dir, profile string) ([]Snippet, error) {
	if dir == "" {
		dir = DefaultDir(profile)
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Snippet{}, nil
		}
		return nil, err
	}
	var out []Snippet
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if !strings.HasSuffix(n, ".yaml") && !strings.HasSuffix(n, ".yml") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			continue
		}
		out = append(out, parse(profile, string(b)))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pin != out[j].Pin {
			return out[i].Pin
		}
		if out[i].LastUsedAt != out[j].LastUsedAt {
			return out[i].LastUsedAt > out[j].LastUsedAt
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func Touch(dir string, s Snippet) error {
	if s.Profile == "" || s.Name == "" {
		return nil
	}
	s.UseCount++
	s.LastUsedAt = time.Now().Format(time.RFC3339)
	_, err := Save(dir, s, true)
	return err
}

func safeName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "-")
	var out []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return "snippet"
	}
	return string(out)
}

func parse(profile, src string) Snippet {
	s := Snippet{Version: 1, Profile: profile, Headers: map[string]string{}}
	lines := strings.Split(src, "\n")
	var inNotes, inReq, inHeaders, inBody bool
	var body []string
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trim := strings.TrimSpace(line)

		if strings.HasPrefix(trim, "notes: |") {
			inNotes = true
			inReq, inHeaders, inBody = false, false, false
			continue
		}
		if strings.HasPrefix(trim, "request:") {
			inReq = true
			inNotes, inBody = false, false
			continue
		}
		if inReq && strings.HasPrefix(trim, "headers:") {
			inHeaders = true
			continue
		}
		if inReq && strings.HasPrefix(trim, "body: |") {
			inBody = true
			body = body[:0]
			continue
		}

		if inNotes {
			if strings.HasPrefix(line, "  ") {
				s.Notes += strings.TrimPrefix(line, "  ") + "\n"
			}
			continue
		}
		if inBody {
			if strings.HasPrefix(line, "    ") {
				body = append(body, strings.TrimPrefix(line, "    "))
			}
			continue
		}

		if idx := strings.Index(trim, ":"); idx > 0 && !inHeaders {
			k := strings.TrimSpace(trim[:idx])
			v := strings.TrimSpace(trim[idx+1:])
			switch k {
			case "name":
				s.Name = v
			case "profile":
				s.Profile = v
			case "createdAt":
				s.CreatedAt = v
			case "lastUsedAt":
				s.LastUsedAt = v
			case "useCount":
				s.UseCount = atoi(v)
			case "successCount":
				s.SuccessCount = atoi(v)
			case "failCount":
				s.FailCount = atoi(v)
			case "avgLatencyMs":
				s.AvgLatencyMs = atoi64(v)
			case "pin":
				s.Pin = (v == "true")
			}
		}

		if inReq {
			if strings.HasPrefix(trim, "method:") {
				s.Method = strings.TrimSpace(strings.TrimPrefix(trim, "method:"))
			}
			if strings.HasPrefix(trim, "path:") {
				s.Path = strings.TrimSpace(strings.TrimPrefix(trim, "path:"))
			}
			if inHeaders {
				if idx := strings.Index(trim, ":"); idx > 0 {
					k := strings.TrimSpace(trim[:idx])
					v := strings.TrimSpace(trim[idx+1:])
					if k != "" && v != "" {
						s.Headers[k] = v
					}
				}
			}
		}
	}
	if len(body) > 0 {
		s.Body = strings.Join(body, "\n")
	}
	return s
}

func atoi(s string) int {
	n := 0
	ok := false
	for _, r := range s {
		if r < '0' || r > '9' {
			if ok {
				break
			}
			continue
		}
		ok = true
		n = n*10 + int(r-'0')
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [32]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func TouchWithResult(dir string, s Snippet, ok bool, latencyMs int64) error {
	if s.Profile == "" || s.Name == "" {
		return nil
	}
	s.UseCount++
	s.LastUsedAt = time.Now().Format(time.RFC3339)
	if ok {
		s.SuccessCount++
	} else {
		s.FailCount++
	}
	if s.UseCount > 0 {
		prev := s.AvgLatencyMs
		if prev == 0 {
			s.AvgLatencyMs = latencyMs
		} else {
			s.AvgLatencyMs = (prev*int64(s.UseCount-1) + latencyMs) / int64(s.UseCount)
		}
	}
	_, err := Save(dir, s, true)
	return err
}

func atoi64(s string) int64 {
	var n int64
	ok := false
	for _, r := range s {
		if r < '0' || r > '9' {
			if ok {
				break
			}
			continue
		}
		ok = true
		n = n*10 + int64(r-'0')
	}
	return n
}

func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [32]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
