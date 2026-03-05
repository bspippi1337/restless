package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

type ProbeResult struct {
	Target     string              `json:"target"`
	FinalURL   string              `json:"final_url"`
	Status     string              `json:"status"`
	StatusCode int                 `json:"status_code"`
	MethodHint []string            `json:"method_hints,omitempty"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Notes      []string            `json:"notes,omitempty"`
}

func normalizeTarget(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return "https://" + s
}

func doReq(method, target string, body io.Reader) (*http.Response, []byte, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	req, err := http.NewRequest(method, target, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "restless/dev (+https://github.com/bspippi1337/restless)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	b, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, b, nil
}

func Probe(target string) (*ProbeResult, error) {
	t := normalizeTarget(target)

	notes := []string{}
	resp, _, err := doReq("HEAD", t, nil)
	if err != nil || resp == nil {
		notes = append(notes, "HEAD failed, trying GET")
		resp, _, err = doReq("GET", t, nil)
		if err != nil {
			return nil, err
		}
	}

	finalURL := ""
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	// OPTIONS -> Allow (best effort)
	methodHints := []string{}
	if u, e := url.Parse(finalURL); e == nil && u.Scheme != "" {
		if oResp, _, oErr := doReq("OPTIONS", finalURL, nil); oErr == nil && oResp != nil {
			allow := oResp.Header.Get("Allow")
			if allow != "" {
				for _, p := range strings.Split(allow, ",") {
					m := strings.TrimSpace(p)
					if m != "" {
						methodHints = append(methodHints, strings.ToUpper(m))
					}
				}
			}
		}
	}

	if len(methodHints) > 0 {
		mset := map[string]bool{}
		out := []string{}
		for _, m := range methodHints {
			if !mset[m] {
				mset[m] = true
				out = append(out, m)
			}
		}
		sort.Strings(out)
		methodHints = out
	}

	return &ProbeResult{
		Target:     target,
		FinalURL:   finalURL,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		MethodHint: methodHints,
		Headers:    resp.Header,
		Notes:      notes,
	}, nil
}

func printJSON(v any, out io.Writer) int {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, "json encode error:", err)
		return 1
	}
	return 0
}

func RunHTTP(method, target string, in io.Reader, out io.Writer) int {
	t := normalizeTarget(target)

	var body io.Reader
	if method == "POST" || method == "PUT" || method == "PATCH" {
		b, _ := io.ReadAll(in)
		if len(bytes.TrimSpace(b)) > 0 {
			body = bytes.NewReader(b)
		}
	}

	resp, b, err := doReq(method, t, body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request error:", err)
		return 1
	}

	fmt.Fprintf(out, "%s %s\n", method, resp.Request.URL.String())
	fmt.Fprintf(out, "Status: %s\n\n", resp.Status)
	for k, vv := range resp.Header {
		for _, v := range vv {
			fmt.Fprintf(out, "%s: %s\n", k, v)
		}
	}
	fmt.Fprintln(out, "")
	_, _ = out.Write(b)
	if len(b) > 0 && b[len(b)-1] != '\n' {
		fmt.Fprintln(out, "")
	}
	return 0
}

func Usage(out io.Writer) {
	fmt.Fprintln(out, "restless")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  restless <domain-or-url>          # default: smart")
	fmt.Fprintln(out, "  restless probe <domain-or-url>")
	fmt.Fprintln(out, "  restless smart <domain-or-url>")
	fmt.Fprintln(out, "  restless simulate <domain-or-url> # alias: smart (for now)")
	fmt.Fprintln(out, "  restless <METHOD> <url>           # raw HTTP (GET/POST/...)")
	fmt.Fprintln(out, "")
}

func Main(args []string, in io.Reader, out io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		Usage(out)
		return 0
	}

	// raw HTTP: METHOD URL
	if len(args) >= 2 && len(args[0]) <= 8 {
		m := strings.ToUpper(args[0])
		switch m {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
			return RunHTTP(m, args[1], in, out)
		}
	}

	cmd := strings.ToLower(args[0])

	// default: treat as target -> smart
	if cmd != "probe" && cmd != "smart" && cmd != "simulate" {
		return Main(append([]string{"smart"}, args...), in, out)
	}

	if len(args) < 2 {
		Usage(out)
		return 2
	}
	target := args[1]

	switch cmd {
	case "probe":
		r, err := Probe(target)
		if err != nil {
			fmt.Fprintln(os.Stderr, "probe error:", err)
			return 1
		}
		return printJSON(r, out)

	case "smart", "simulate":
		r, err := Probe(target)
		if err != nil {
			fmt.Fprintln(os.Stderr, "smart error:", err)
			return 1
		}
		if rc := printJSON(r, out); rc != 0 {
			return rc
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Next:")
		fmt.Fprintln(out, "  restless GET  "+normalizeTarget(target))
		fmt.Fprintln(out, "  restless probe "+target)
		return 0
	}

	Usage(out)
	return 2
}
