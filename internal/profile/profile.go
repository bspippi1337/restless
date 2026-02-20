package profile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Profile struct {
	Name     string
	Path     string
	BaseURLs []string
	AuthType string
	AuthEnv  string
	Defaults map[string]string
	TimeoutS int
	Endpoints []Endpoint
}

type Endpoint struct {
	Method string
	Path   string
	Score  float64
}

func DefaultProfileDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".config", "restless", "profiles")
}

func Load(profileDir, name string) (Profile, error) {
	if name == "" {
		return Profile{}, errors.New("empty profile name")
	}
	if profileDir == "" {
		profileDir = DefaultProfileDir()
	}
	p := filepath.Join(profileDir, name+".yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		p2 := filepath.Join(profileDir, name+".yml")
		b2, err2 := os.ReadFile(p2)
		if err2 != nil {
			return Profile{}, err
		}
		p = p2
		b = b2
	}

	pr := Profile{
		Name:     name,
		Path:     p,
		Defaults: map[string]string{},
		TimeoutS: 20,
	}

	lines := strings.Split(string(b), "\n")
	var inBase, inAuth, inToken, inDefaults, inHeaders, inEndpoints bool
	var curEp *Endpoint

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trim := strings.TrimSpace(line)

		switch trim {
		case "baseUrls:":
			inBase, inAuth, inToken, inDefaults, inHeaders, inEndpoints = true, false, false, false, false, false
			continue
		case "auth:":
			inBase, inAuth, inToken, inDefaults, inHeaders, inEndpoints = false, true, false, false, false, false
			continue
		case "token:":
			if inAuth { inToken = true }
			continue
		case "defaults:":
			inBase, inAuth, inToken, inDefaults, inHeaders, inEndpoints = false, false, false, true, false, false
			continue
		case "headers:":
			if inDefaults { inHeaders = true }
			continue
		case "endpoints:":
			inBase, inAuth, inToken, inDefaults, inHeaders, inEndpoints = false, false, false, false, false, true
			continue
		}

		if strings.HasPrefix(trim, "timeoutSeconds:") && inDefaults {
			pr.TimeoutS = atoiSafe(afterColon(trim), 20)
			continue
		}

		if inBase {
			if strings.HasPrefix(trim, "- ") {
				pr.BaseURLs = append(pr.BaseURLs, strings.TrimSpace(strings.TrimPrefix(trim, "- ")))
			}
			continue
		}

		if inAuth && strings.HasPrefix(trim, "type:") {
			pr.AuthType = strings.TrimSpace(afterColon(trim))
			continue
		}
		if inToken && strings.HasPrefix(trim, "envVar:") {
			pr.AuthEnv = strings.TrimSpace(afterColon(trim))
			continue
		}

		if inHeaders {
			if trim == "" || strings.HasSuffix(trim, ":") { continue }
			if idx := strings.Index(trim, ":"); idx > 0 {
				k := strings.TrimSpace(trim[:idx])
				v := strings.TrimSpace(trim[idx+1:])
				if k != "" && v != "" { pr.Defaults[k] = v }
			}
			continue
		}

		if inEndpoints {
			if strings.HasPrefix(trim, "- method:") {
				ep := Endpoint{Method: strings.TrimSpace(strings.TrimPrefix(trim, "- method:"))}
				pr.Endpoints = append(pr.Endpoints, ep)
				curEp = &pr.Endpoints[len(pr.Endpoints)-1]
				continue
			}
			if curEp != nil && strings.HasPrefix(trim, "path:") {
				curEp.Path = strings.TrimSpace(afterColon(trim))
				continue
			}
			if curEp != nil && strings.HasPrefix(trim, "score:") {
				curEp.Score = atofSafe(afterColon(trim), 0)
				continue
			}
		}
	}

	if len(pr.BaseURLs) == 0 {
		pr.BaseURLs = []string{"https://api.example.com"}
	}
	if pr.AuthEnv == "" { pr.AuthEnv = "RESTLESS_TOKEN" }
	if pr.AuthType == "" { pr.AuthType = "bearer" }

	return pr, nil
}

func afterColon(s string) string {
	if i := strings.Index(s, ":"); i >= 0 { return strings.TrimSpace(s[i+1:]) }
	return ""
}

func atoiSafe(s string, def int) int {
	n := 0; ok := false
	for _, r := range s {
		if r < '0' || r > '9' { if ok { break }; continue }
		ok = true; n = n*10 + int(r-'0')
	}
	if !ok { return def }
	return n
}

func atofSafe(s string, def float64) float64 {
	s = strings.TrimSpace(s)
	if s == "" { return def }
	sign := 1.0
	if strings.HasPrefix(s, "-") { sign = -1.0; s = strings.TrimPrefix(s, "-") }
	parts := strings.SplitN(s, ".", 2)
	i := float64(atoiSafe(parts[0], 0))
	if len(parts) == 1 { return sign * i }
	frac := 0.0; div := 1.0
	for _, r := range parts[1] {
		if r < '0' || r > '9' { break }
		frac = frac*10 + float64(r-'0'); div *= 10
	}
	return sign * (i + frac/div)
}
