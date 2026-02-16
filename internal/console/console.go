package console

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/history"
	"github.com/bspippi1337/restless/internal/httpclient"
	"github.com/bspippi1337/restless/internal/profile"
	"github.com/bspippi1337/restless/internal/snippets"
)

type Options struct {
	ProfileName string
	ProfileDir  string
	SnippetDir  string
	BaseURL     string
}

type session struct {
	pr      profile.Profile
	opt     Options
	baseURL string
	cur     httpclient.Request
	history []httpclient.Request
}

func Run(opt Options) error {
	pr, err := profile.Load(opt.ProfileDir, opt.ProfileName)
	if err != nil { return fmt.Errorf("load profile: %w", err) }
	s := &session{pr: pr, opt: opt}
	if opt.BaseURL != "" { s.baseURL = opt.BaseURL } else { s.baseURL = pr.BaseURLs[0] }

	s.cur = httpclient.Request{
		Method:  "GET",
		BaseURL: s.baseURL,
		Path:    "/",
		Headers: cloneMap(pr.Defaults),
		Query:   map[string]string{},
	}

	printBanner(pr.Name, s.baseURL)
	fmt.Println("Type 'help' for commands. Tip: suggest -> run -> save <name>.")

	in := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("restless> ")
		if !in.Scan() { fmt.Println(); return nil }
		line := strings.TrimSpace(in.Text())
		if line == "" { continue }
		parts := splitArgs(line)
		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		switch cmd {
		case "exit", "quit", "q":
			return nil
		case "help", "?":
			printHelp()
		case "status":
			s.printStatus()
		case "base":
			s.cmdBase(args)
		case "endpoints", "eps":
			s.cmdEndpoints()
		case "suggest":
			s.cmdSuggest()
		case "pick":
			s.cmdPick(args)
		case "set":
			s.cmdSet(args)
		case "header":
			s.cmdHeader(args)
		case "query":
			s.cmdQuery(args)
		case "body":
			s.cmdBody()
		case "show":
			s.cmdShow()
		case "run":
			s.cmdRun()
		case "save":
			s.cmdSave(args)
		case "snippets", "snips":
			s.cmdSnips()
		case "use":
			s.cmdUse(args)
		case "pin":
			s.cmdPin(args, true)
		case "unpin":
			s.cmdPin(args, false)
		case "export":
			s.cmdExport(args)
		case "history":
			s.cmdHistory()
		default:
			fmt.Println("Unknown command:", cmd, "(try: help)")
		}
	}
}

func printBanner(profileName, baseURL string) {
	fmt.Println("┌──────────────────────────────────────────────┐")
	fmt.Println("│ Restless Console (v1)                         │")
	fmt.Printf("│ Profile: %-35s│\n", clip(profileName, 35))
	fmt.Printf("│ BaseURL : %-35s│\n", clip(baseURL, 35))
	fmt.Println("└──────────────────────────────────────────────┘")
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  help                      show this help")
	fmt.Println("  status                    show active profile/base/request")
	fmt.Println("  base [url]                show/set base URL")
	fmt.Println("  endpoints                 list discovered endpoints (numbered)")
	fmt.Println("  suggest                   pick best endpoint into current request")
	fmt.Println("  pick <n>                  set current request to endpoint #n")
	fmt.Println("  set <METHOD> <PATH>       manually set method+path")
	fmt.Println("  header <K> <V>            set header")
	fmt.Println("  query <K> <V>             set query param")
	fmt.Println("  body                      enter JSON body (multi-line, end with '.')")
	fmt.Println("  show                      show current request")
	fmt.Println("  run                       execute current request")
	fmt.Println("  save <name>               save current request as snippet")
	fmt.Println("  snippets                  list snippets (pinned first)")
	fmt.Println("  use <name>                load+run snippet")
	fmt.Println("  pin <name> / unpin <name> pin/unpin snippet")
	fmt.Println("  export <curl|httpie>      export current request")
	fmt.Println("  history                   show recent requests")
	fmt.Println("  quit                      exit")
}

func (s *session) printStatus() {
	fmt.Println("Profile:", s.pr.Name)
	fmt.Println("Profile file:", s.pr.Path)
	fmt.Println("BaseURL:", s.baseURL)
	fmt.Println("Current:", s.cur.Method, s.cur.Path)
	fmt.Println("Endpoints:", len(s.pr.Endpoints))
}

func (s *session) cmdBase(args []string) {
	if len(args) == 0 { fmt.Println("BaseURL:", s.baseURL); return }
	s.baseURL = args[0]
	s.cur.BaseURL = s.baseURL
	fmt.Println("BaseURL set:", s.baseURL)
}

func (s *session) cmdEndpoints() {
	if len(s.pr.Endpoints) == 0 { fmt.Println("(no endpoints in profile)"); return }
	type row struct{ idx int; m, p string; sc float64 }
	rows := make([]row, 0, len(s.pr.Endpoints))
	for i, ep := range s.pr.Endpoints {
		rows = append(rows, row{idx: i + 1, m: ep.Method, p: ep.Path, sc: ep.Score})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].sc > rows[j].sc })
	for _, r := range rows {
		fmt.Printf("%3d) %-6s %-40s score=%.2f\n", r.idx, r.m, clip(r.p, 40), r.sc)
	}
}

func (s *session) cmdSuggest() {
	if len(s.pr.Endpoints) == 0 { fmt.Println("No endpoints in profile."); return }
	best := s.pr.Endpoints[0]
	for _, ep := range s.pr.Endpoints[1:] { if ep.Score > best.Score { best = ep } }
	s.cur.Method = strings.ToUpper(best.Method)
	s.cur.Path = best.Path
	s.cur.BaseURL = s.baseURL
	if s.cur.Headers == nil { s.cur.Headers = cloneMap(s.pr.Defaults) }
	if s.cur.Query == nil { s.cur.Query = map[string]string{} }
	fmt.Println("Suggested:", s.cur.Method, s.cur.Path)
}

func (s *session) cmdPick(args []string) {
	if len(args) < 1 { fmt.Println("usage: pick <n>"); return }
	n := atoi(args[0])
	if n <= 0 || n > len(s.pr.Endpoints) { fmt.Println("invalid endpoint index"); return }
	ep := s.pr.Endpoints[n-1]
	s.cur.Method = strings.ToUpper(ep.Method)
	s.cur.Path = ep.Path
	s.cur.BaseURL = s.baseURL
	if s.cur.Headers == nil { s.cur.Headers = cloneMap(s.pr.Defaults) }
	if s.cur.Query == nil { s.cur.Query = map[string]string{} }
	fmt.Println("Picked:", s.cur.Method, s.cur.Path)
}

func (s *session) cmdSet(args []string) {
	if len(args) < 2 { fmt.Println("usage: set <METHOD> <PATH>"); return }
	s.cur.Method = strings.ToUpper(args[0])
	s.cur.Path = args[1]
	fmt.Println("Current set:", s.cur.Method, s.cur.Path)
}

func (s *session) cmdHeader(args []string) {
	if len(args) < 2 { fmt.Println("usage: header <K> <V>"); return }
	if s.cur.Headers == nil { s.cur.Headers = map[string]string{} }
	s.cur.Headers[args[0]] = strings.Join(args[1:], " ")
	fmt.Println("Header set:", args[0])
}

func (s *session) cmdQuery(args []string) {
	if len(args) < 2 { fmt.Println("usage: query <K> <V>"); return }
	if s.cur.Query == nil { s.cur.Query = map[string]string{} }
	s.cur.Query[args[0]] = strings.Join(args[1:], " ")
	fmt.Println("Query set:", args[0])
}

func (s *session) cmdBody() {
	fmt.Println("Enter JSON body. End with a single dot '.' on its own line.")
	var lines []string
	in := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("... ")
		if !in.Scan() { break }
		t := in.Text()
		if strings.TrimSpace(t) == "." { break }
		lines = append(lines, t)
	}
	s.cur.Body = []byte(strings.Join(lines, "\n"))
	fmt.Println("Body set (bytes):", len(s.cur.Body))
}

func (s *session) cmdShow() {
	fmt.Println("Method:", s.cur.Method)
	fmt.Println("BaseURL:", s.cur.BaseURL)
	fmt.Println("Path:", s.cur.Path)
	if len(s.cur.Query) > 0 {
		fmt.Println("Query:")
		for k, v := range s.cur.Query { fmt.Printf("  %s=%s\n", k, v) }
	}
	fmt.Println("Headers:")
	keys := make([]string, 0, len(s.cur.Headers))
	for k := range s.cur.Headers { keys = append(keys, k) }
	sort.Strings(keys)
	for _, k := range keys { fmt.Printf("  %s: %s\n", k, s.cur.Headers[k]) }
	if len(s.cur.Body) > 0 {
		fmt.Println("Body:")
		fmt.Println(string(s.cur.Body))
	}
}

func (s *session) cmdRun() {
	timeout := time.Duration(s.pr.TimeoutS) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if strings.ToLower(s.pr.AuthType) == "bearer" {
		if s.pr.AuthEnv != "" && s.cur.Headers != nil && s.cur.Headers["Authorization"] == "" {
			if tok := os.Getenv(s.pr.AuthEnv); tok != "" {
				s.cur.Headers["Authorization"] = "Bearer " + tok
			}
		}
	}

	res, err := httpclient.Do(ctx, s.cur)
	s.history = append([]httpclient.Request{s.cur}, s.history...)
	if len(s.history) > 25 { s.history = s.history[:25] }
	if err != nil { fmt.Println("❌ error:", err); return }

	 _ = history.Append(s.pr.Name, history.Entry{Profile: s.pr.Name, Method: s.cur.Method, Path: s.cur.Path, BaseURL: s.cur.BaseURL, StatusCode: res.StatusCode, LatencyMs: res.LatencyMs, OK: res.StatusCode >= 200 && res.StatusCode < 400})
fmt.Printf("✅ %s (%d) %dms\n", res.Status, res.StatusCode, res.LatencyMs)
	fmt.Println(httpclient.Redact(string(httpclient.PrettyJSON(res.Body))))
}

func (s *session) cmdSave(args []string) {
	if len(args) < 1 { fmt.Println("usage: save <name>"); return }
	name := args[0]
	sn := snippets.Snippet{
		Name:    name,
		Profile: s.pr.Name,
		Method:  s.cur.Method,
		Path:    s.cur.Path,
		Headers: cloneMap(s.cur.Headers),
		Notes:   "Saved from console.",
	}
	if len(s.cur.Body) > 0 { sn.Body = string(s.cur.Body) }
	path, err := snippets.Save(s.opt.SnippetDir, sn, false)
	if err != nil { fmt.Println("save error:", err); return }
	fmt.Println("💾 Saved:", path)
}

func (s *session) cmdSnips() {
	list, err := snippets.List(s.opt.SnippetDir, s.pr.Name)
	if err != nil { fmt.Println("snippets error:", err); return }
	if len(list) == 0 { fmt.Println("(no snippets yet)"); return }
	for _, sn := range list {
		pin := " "; if sn.Pin { pin = "★" }
		fmt.Printf("%s %-16s %-6s %-36s used=%d last=%s\n", pin, clip(sn.Name, 16), strings.ToUpper(sn.Method), clip(sn.Path, 36), sn.UseCount, clip(sn.LastUsedAt, 20))
	}
}

func (s *session) cmdUse(args []string) {
	if len(args) < 1 { fmt.Println("usage: use <name>"); return }
	name := args[0]
	sn, err := snippets.Load(s.opt.SnippetDir, s.pr.Name, name)
	if err != nil { fmt.Println("load error:", err); return }

	headers := cloneMap(s.pr.Defaults)
	for k, v := range sn.Headers { headers[k] = v }
	if strings.ToLower(s.pr.AuthType) == "bearer" && headers["Authorization"] == "" {
		if tok := os.Getenv(s.pr.AuthEnv); tok != "" { headers["Authorization"] = "Bearer " + tok }
	}

	req := httpclient.Request{Method: strings.ToUpper(sn.Method), BaseURL: s.baseURL, Path: sn.Path, Headers: headers, Query: map[string]string{}, Body: []byte(sn.Body)}
	timeout := time.Duration(s.pr.TimeoutS) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := httpclient.Do(ctx, req)
	ok := false
	if err != nil {
		fmt.Println("❌ error:", err)
	} else {
		ok = res.StatusCode >= 200 && res.StatusCode < 400
		fmt.Printf("✅ %s (%d) %dms\n", res.Status, res.StatusCode, res.LatencyMs)
		fmt.Println(httpclient.Redact(string(httpclient.PrettyJSON(res.Body))))
	}
	_ = history.Append(s.pr.Name, history.Entry{Profile: s.pr.Name, Name: sn.Name, Method: req.Method, Path: req.Path, BaseURL: req.BaseURL, StatusCode: res.StatusCode, LatencyMs: res.LatencyMs, OK: ok})
	_ = snippets.TouchWithResult(s.opt.SnippetDir, sn, ok, res.LatencyMs)
}

func (s *session) cmdPinfunc (s *session) cmdPin(args []string, pin bool) {
	if len(args) < 1 { fmt.Println("usage: pin <name>"); return }
	name := args[0]
	sn, err := snippets.Load(s.opt.SnippetDir, s.pr.Name, name)
	if err != nil { fmt.Println("load error:", err); return }
	sn.Pin = pin
	_, err = snippets.Save(s.opt.SnippetDir, sn, true)
	if err != nil { fmt.Println("save error:", err); return }
	if pin { fmt.Println("★ pinned", name) } else { fmt.Println("unpinned", name) }
}

func (s *session) cmdExport(args []string) {
	if len(args) < 1 { fmt.Println("usage: export <curl|httpie>"); return }
	format := strings.ToLower(args[0])
	full, _ := httpclient.BuildURL(s.cur.BaseURL, s.cur.Path, s.cur.Query)
	switch format {
	case "curl":
		fmt.Println(toCurl(full, s.cur))
	case "httpie":
		fmt.Println(toHTTPie(full, s.cur))
	default:
		fmt.Println("unknown export format")
	}
}

func (s *session) cmdHistory() {
	if len(s.history) == 0 { fmt.Println("(empty)"); return }
	for i, r := range s.history { fmt.Printf("%2d) %-6s %s\n", i+1, r.Method, r.Path) }
}

func toCurl(url string, r httpclient.Request) string {
	var b strings.Builder
	b.WriteString("curl -i '"); b.WriteString(url); b.WriteString("' -X "); b.WriteString(strings.ToUpper(r.Method))
	for k, v := range r.Headers {
		b.WriteString(" -H '"); b.WriteString(k); b.WriteString(": "); b.WriteString(v); b.WriteString("'")
	}
	if len(r.Body) > 0 {
		b.WriteString(" --data '"); b.WriteString(strings.ReplaceAll(string(r.Body), "'", "'\"'\"'")); b.WriteString("'")
	}
	return b.String()
}

func toHTTPie(url string, r httpclient.Request) string {
	var b strings.Builder
	b.WriteString("http "); b.WriteString(strings.ToUpper(r.Method)); b.WriteString(" '"); b.WriteString(url); b.WriteString("'")
	for k, v := range r.Headers {
		b.WriteString(" "); b.WriteString(k); b.WriteString(":'"); b.WriteString(v); b.WriteString("'")
	}
	if len(r.Body) > 0 {
		b.WriteString(" <<< '"); b.WriteString(strings.ReplaceAll(string(r.Body), "'", "'\"'\"'")); b.WriteString("'")
	}
	return b.String()
}

func splitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	inQ := false
	esc := false
	for _, r := range s {
		if esc { cur.WriteRune(r); esc = false; continue }
		if r == '\\' { esc = true; continue }
		if r == '"' { inQ = !inQ; continue }
		if !inQ && (r == ' ' || r == '\t') {
			if cur.Len() > 0 { out = append(out, cur.String()); cur.Reset() }
			continue
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 { out = append(out, cur.String()) }
	return out
}

func cloneMap(m map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range m { out[k] = v }
	return out
}

func atoi(s string) int {
	n := 0; ok := false
	for _, r := range s {
		if r < '0' || r > '9' { if ok { break }; continue }
		ok = true; n = n*10 + int(r-'0')
	}
	return n
}

func clip(s string, n int) string {
	if len(s) <= n { return s }
	if n <= 1 { return s[:1] }
	return s[:n-1] + "…"
}
