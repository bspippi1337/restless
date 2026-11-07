package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Visit struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Status int    `json:"status"`
	At     string `json:"at"`
}

type State struct {
	mu     sync.Mutex
	Visits []Visit
	Seen   map[string]bool
}

func main() {
	st := &State{Seen: map[string]bool{}}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("web")))

	mux.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Method  string            `json:"method"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
			Body    any               `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		method := strings.ToUpper(strings.TrimSpace(req.Method))
		if method == "" {
			method = "GET"
		}

		u, err := url.Parse(req.URL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}

		var bodyReader io.Reader
		if req.Body != nil && method != "GET" && method != "HEAD" {
			b, _ := json.Marshal(req.Body)
			bodyReader = bytes.NewReader(b)
		}

		outReq, _ := http.NewRequest(method, req.URL, bodyReader)
		for k, v := range req.Headers {
			if strings.TrimSpace(k) == "" {
				continue
			}
			outReq.Header.Set(k, v)
		}
		if outReq.Header.Get("Accept") == "" {
			outReq.Header.Set("Accept", "application/json, */*")
		}
		if bodyReader != nil && outReq.Header.Get("Content-Type") == "" {
			outReq.Header.Set("Content-Type", "application/json")
		}

		client := &http.Client{Timeout: 25 * time.Second}
		resp, err := client.Do(outReq)
		if err != nil {
			http.Error(w, "request failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		ct := resp.Header.Get("Content-Type")

		st.mu.Lock()
		st.Visits = append(st.Visits, Visit{
			Method: method,
			URL:    req.URL,
			Status: resp.StatusCode,
			At:     time.Now().Format(time.RFC3339),
		})
		st.Seen[req.URL] = true
		if len(st.Visits) > 200 {
			st.Visits = st.Visits[len(st.Visits)-200:]
		}
		st.mu.Unlock()

		hints := buildHints(req.URL, ct, raw)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  resp.StatusCode,
			"headers": resp.Header,
			"body":    tryParseJSON(raw),
			"raw":     string(raw),
			"hints":   hints,
		})
	})

	mux.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		st.mu.Lock()
		defer st.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st.Visits)
	})

	addr := "127.0.0.1:8787"
	if v := strings.TrimSpace(os.Getenv("APINAV_ADDR")); v != "" {
		addr = v
	}

	fmt.Println("API Navigator GUI running on http://" + addr)
	_ = http.ListenAndServe(addr, mux)
}

func tryParseJSON(b []byte) any {
	var v any
	if json.Unmarshal(b, &v) == nil {
		return v
	}
	return nil
}

func buildHints(currentURL, contentType string, raw []byte) []string {
	hints := []string{}

	reURL := regexp.MustCompile(`https?://[^\s"'<>]+`)
	for _, m := range reURL.FindAllString(string(raw), 50) {
		hints = appendUnique(hints, m, 30)
	}

	var v any
	if strings.Contains(strings.ToLower(contentType), "json") && json.Unmarshal(raw, &v) == nil {
		switch t := v.(type) {
		case map[string]any:
			hints = append(hints, suggestFromObject(currentURL, t)...)
		case []any:
			if len(t) > 0 {
				if obj, ok := t[0].(map[string]any); ok {
					hints = append(hints, suggestFromObject(currentURL, obj)...)
				}
			}
		}
	}

	if u, err := url.Parse(currentURL); err == nil {
		parent := *u
		parent.Path = path.Dir(u.Path)
		if parent.Path != "." && parent.Path != "/" {
			hints = appendUnique(hints, parent.String(), 30)
		}
	}

	return uniqueTrim(hints, 30)
}

func suggestFromObject(currentURL string, obj map[string]any) []string {
	out := []string{}
	u, err := url.Parse(currentURL)
	if err != nil {
		return out
	}

	for k, v := range obj {
		kl := strings.ToLower(k)

		if kl == "url" || kl == "href" || kl == "link" || kl == "self" {
			if s, ok := v.(string); ok && strings.HasPrefix(s, "http") {
				out = append(out, s)
			}
		}

		if kl == "id" || strings.HasSuffix(kl, "_id") {
			switch vv := v.(type) {
			case string:
				out = append(out, joinPath(u, vv))
			case float64:
				out = append(out, joinPath(u, fmt.Sprintf("%.0f", vv)))
			}
		}

		if s, ok := v.(string); ok {
			if strings.HasPrefix(s, "/") {
				uu := *u
				uu.Path = s
				uu.RawQuery = ""
				out = append(out, uu.String())
			}
		}

		switch v.(type) {
		case string, float64, bool:
			q := u.Query()
			if _, exists := q[k]; !exists && len(q) < 3 {
				uu := *u
				q.Set(k, fmt.Sprintf("%v", v))
				uu.RawQuery = q.Encode()
				out = append(out, uu.String())
			}
		}
	}

	return uniqueTrim(out, 20)
}

func joinPath(u *url.URL, segment string) string {
	uu := *u
	uu.Path = strings.TrimRight(u.Path, "/") + "/" + url.PathEscape(segment)
	uu.RawQuery = ""
	return uu.String()
}

func appendUnique(list []string, val string, max int) []string {
	for _, x := range list {
		if x == val {
			return list
		}
	}
	if len(list) >= max {
		return list
	}
	return append(list, val)
}

func uniqueTrim(in []string, max int) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
		if len(out) >= max {
			break
		}
	}
	return out
}
