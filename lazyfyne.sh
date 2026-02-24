#!/usr/bin/env bash
set -euo pipefail

APP_NAME="${APP_NAME:-lazyfyne}"
APP_DIR="${APP_DIR:-$APP_NAME}"
BIN_NAME="${BIN_NAME:-$APP_NAME}"

say() { printf "\033[1;36m==>\033[0m %s\n" "$*"; }
warn() { printf "\033[1;33m!!\033[0m %s\n" "$*"; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing required command: $1" >&2; exit 1; }
}

need_cmd go

# Optional: install Linux build deps for Fyne (only if INSTALL_DEPS=1 and apt exists)
install_deps_if_requested() {
  if [[ "${INSTALL_DEPS:-0}" != "1" ]]; then
    return 0
  fi
  if ! command -v apt-get >/dev/null 2>&1; then
    warn "INSTALL_DEPS=1 but apt-get not found. Skipping system deps."
    return 0
  fi

  say "Installing Linux system deps for Fyne (via apt-get)"
  sudo apt-get update -y
  sudo apt-get install -y \
    build-essential pkg-config \
    libgl1-mesa-dev xorg-dev \
    libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev \
    libxxf86vm-dev libasound2-dev
}

install_deps_if_requested

say "Creating project in: ${APP_DIR}"
mkdir -p "${APP_DIR}"
cd "${APP_DIR}"

if [[ ! -f go.mod ]]; then
  say "Initializing go module"
  go mod init "${APP_NAME}" >/dev/null 2>&1 || true
fi

say "Writing main.go (native Fyne GUI)"
cat > main.go <<'EOF'
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Engine struct {
	Client *http.Client
}

func NewEngine() *Engine {
	return &Engine{
		Client: &http.Client{Timeout: 25 * time.Second},
	}
}

func (e *Engine) Request(method, target string, headers map[string]string) (status string, body string, hints []string, err error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "GET"
	}
	target = strings.TrimSpace(target)

	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		return "", "", nil, err
	}
	// Sensible defaults
	req.Header.Set("Accept", "application/json, */*")
	for k, v := range headers {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		req.Header.Set(k, v)
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB
	body = string(b)
	status = resp.Status

	hints = buildHints(target, resp.Header.Get("Content-Type"), b)
	return status, body, hints, nil
}

func buildHints(currentURL, contentType string, raw []byte) []string {
	var out []string

	// 1) Extract absolute URLs from raw response
	reURL := regexp.MustCompile(`https?://[^\s"'<>]+`)
	for _, m := range reURL.FindAllString(string(raw), 40) {
		out = appendUnique(out, m, 30)
	}

	// 2) Extract obvious relative links in JSON fields (simple heuristic)
	if strings.Contains(strings.ToLower(contentType), "json") {
		var v any
		if json.Unmarshal(raw, &v) == nil {
			if obj, ok := v.(map[string]any); ok {
				out = append(out, suggestFromObject(currentURL, obj)...)
			} else if arr, ok := v.([]any); ok && len(arr) > 0 {
				if obj, ok := arr[0].(map[string]any); ok {
					out = append(out, suggestFromObject(currentURL, obj)...)
				}
			}
		}
	}

	// 3) Parent path hint
	if u, err := url.Parse(currentURL); err == nil {
		parent := *u
		parent.RawQuery = ""
		if parent.Path != "" && parent.Path != "/" {
			// trim last segment
			p := parent.Path
			if strings.HasSuffix(p, "/") {
				p = strings.TrimSuffix(p, "/")
			}
			if i := strings.LastIndex(p, "/"); i > 0 {
				parent.Path = p[:i]
				out = appendUnique(out, parent.String(), 30)
			}
		}
	}

	return uniqueTrim(out, 30)
}

func suggestFromObject(currentURL string, obj map[string]any) []string {
	u, err := url.Parse(currentURL)
	if err != nil {
		return nil
	}
	var out []string

	for k, v := range obj {
		kl := strings.ToLower(k)

		// Common link-ish keys
		if kl == "url" || kl == "href" || kl == "link" || kl == "self" {
			if s, ok := v.(string); ok && strings.HasPrefix(s, "http") {
				out = append(out, s)
			}
		}

		// Relative path values
		if s, ok := v.(string); ok && strings.HasPrefix(s, "/") {
			uu := *u
			uu.Path = s
			uu.RawQuery = ""
			out = append(out, uu.String())
		}

		// id fields -> propose /{id}
		if kl == "id" || strings.HasSuffix(kl, "_id") {
			switch vv := v.(type) {
			case string:
				out = append(out, joinPath(u, vv))
			case float64:
				out = append(out, joinPath(u, strings.TrimSuffix(strings.TrimSuffix(
					strings.TrimRight(strings.TrimRight(strings.TrimRight(
						formatFloat(vv), "0"), "."), "0"), "."), ".")))
			}
		}
	}

	return uniqueTrim(out, 20)
}

func formatFloat(f float64) string {
	// enough for IDs
	b, _ := json.Marshal(f)
	return string(b)
}

func joinPath(u *url.URL, segment string) string {
	uu := *u
	uu.RawQuery = ""
	uu.Path = strings.TrimRight(u.Path, "/") + "/" + url.PathEscape(segment)
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

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("LazyFyne API Navigator")
	w.Resize(fyne.NewSize(1180, 720))

	engine := NewEngine()

	// --- Inputs
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://api.example.com/users")

	methodSelect := widget.NewSelect([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, nil)
	methodSelect.SetSelected("GET")

	headersEntry := widget.NewMultiLineEntry()
	headersEntry.SetPlaceHolder(`{"Accept":"application/json","Authorization":"Bearer ...optional..."}`)
	headersEntry.SetText(`{"Accept":"application/json"}`)
	headersEntry.Wrapping = fyne.TextWrapWord

	// --- Outputs
	statusLabel := widget.NewLabel("HTTP -")
	responseBox := widget.NewMultiLineEntry()
	responseBox.Wrapping = fyne.TextWrapWord

	var hints []string
	hintsList := widget.NewList(
		func() int { return len(hints) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(hints[i])
		},
	)
	hintsList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(hints) {
			urlEntry.SetText(hints[id])
			methodSelect.SetSelected("GET")
		}
	}

	sendBtn := widget.NewButton("Send", func() {
		target := strings.TrimSpace(urlEntry.Text)
		if target == "" {
			responseBox.SetText("Please enter a URL.")
			return
		}

		// Parse headers JSON
		hdrs := map[string]string{}
		if strings.TrimSpace(headersEntry.Text) != "" {
			if err := json.Unmarshal([]byte(headersEntry.Text), &hdrs); err != nil {
				responseBox.SetText("Headers JSON invalid:\n" + err.Error())
				return
			}
		}

		status, body, newHints, err := engine.Request(methodSelect.Selected, target, hdrs)
		if err != nil {
			statusLabel.SetText("HTTP error")
			responseBox.SetText(err.Error())
			hints = nil
			hintsList.Refresh()
			return
		}
		statusLabel.SetText(status)

		// Pretty print JSON if possible
		var pretty any
		if json.Unmarshal([]byte(body), &pretty) == nil {
			b, _ := json.MarshalIndent(pretty, "", "  ")
			responseBox.SetText(string(b))
		} else {
			responseBox.SetText(body)
		}

		hints = newHints
		hintsList.Refresh()
	})

	// Layout
	left := container.NewVBox(
		widget.NewLabelWithStyle("Request", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("URL"),
		urlEntry,
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Method"), methodSelect),
			container.NewVBox(widget.NewLabel("HTTP Status"), statusLabel),
		),
		widget.NewLabel("Headers (JSON)"),
		container.NewVScroll(headersEntry),
		sendBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Next choices", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewVScroll(hintsList),
	)

	right := container.NewVBox(
		widget.NewLabelWithStyle("Response", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewVScroll(responseBox),
	)

	split := container.NewHSplit(left, right)
	split.Offset = 0.36

	w.SetContent(split)
	w.ShowAndRun()
}
EOF

say "Fetching Fyne"
go get fyne.io/fyne/v2 >/dev/null

say "Tidying"
go mod tidy >/dev/null

say "Formatting (soft)"
gofmt -w main.go >/dev/null 2>&1 || true

say "Building"
go build -o "${BIN_NAME}"

say "Launching ./${BIN_NAME}"
./"${BIN_NAME}"
