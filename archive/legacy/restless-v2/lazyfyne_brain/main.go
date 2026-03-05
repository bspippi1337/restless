package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Engine struct {
	Client   *http.Client
	LastJSON any
	LastURL  string
}

func NewEngine() *Engine {
	return &Engine{Client: &http.Client{Timeout: 25 * time.Second}}
}

func (e *Engine) Request(method, target string) (string, string, []string, []string, error) {
	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		return "", "", nil, nil, err
	}
	req.Header.Set("Accept", "application/json, */*")

	resp, err := e.Client.Do(req)
	if err != nil {
		return "", "", nil, nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	body := string(bodyBytes)

	var parsed any
	var smart []string

	if json.Unmarshal(bodyBytes, &parsed) == nil {
		e.LastJSON = parsed
		e.LastURL = target
		smart = extractSmart(parsed, target)
	}

	rawLinks := extractRawLinks(body)

	return resp.Status, body, rawLinks, smart, nil
}

func extractRawLinks(body string) []string {
	var out []string
	reURL := regexp.MustCompile(`https?://[^\s"'<>]+`)
	for _, m := range reURL.FindAllString(body, 20) {
		out = append(out, m)
	}
	return unique(out)
}

func extractSmart(v any, baseURL string) []string {
	var out []string
	u, _ := url.Parse(baseURL)

	switch t := v.(type) {

	case map[string]any:
		for k, val := range t {
			keyLower := strings.ToLower(k)

			if keyLower == "id" || strings.HasSuffix(keyLower, "_id") {
				switch vv := val.(type) {
				case float64:
					out = append(out, joinPath(u, strconv.Itoa(int(vv))))
				case string:
					out = append(out, joinPath(u, vv))
				}
			}

			switch vv := val.(type) {
			case string:
				out = append(out, addQuery(u, k, vv))
			case float64:
				out = append(out, addQuery(u, k, strconv.Itoa(int(vv))))
			case bool:
				out = append(out, addQuery(u, k, strconv.FormatBool(vv)))
			}
		}

	case []any:
		if len(t) > 0 {
			return extractSmart(t[0], baseURL)
		}
	}

	return unique(out)
}

func joinPath(u *url.URL, segment string) string {
	uu := *u
	uu.RawQuery = ""
	uu.Path = strings.TrimRight(u.Path, "/") + "/" + segment
	return uu.String()
}

func addQuery(u *url.URL, key, value string) string {
	uu := *u
	q := uu.Query()
	q.Set(key, value)
	uu.RawQuery = q.Encode()
	return uu.String()
}

func unique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("LazyFyne Brain++")
	w.Resize(fyne.NewSize(1300, 800))

	engine := NewEngine()

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://api.example.com")

	methodSelect := widget.NewSelect([]string{"GET", "POST", "PUT", "DELETE"}, nil)
	methodSelect.SetSelected("GET")

	responseBox := widget.NewMultiLineEntry()
	responseBox.Wrapping = fyne.TextWrapWord

	var rawLinks []string
	var smart []string

	rawList := widget.NewList(
		func() int { return len(rawLinks) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(rawLinks[i])
		},
	)

	smartList := widget.NewList(
		func() int { return len(smart) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(smart[i])
		},
	)

	smartList.OnSelected = func(id widget.ListItemID) {
		urlEntry.SetText(smart[id])
		methodSelect.SetSelected("GET")
	}

	sendBtn := widget.NewButton("Send", func() {
		status, body, newRaw, newSmart, err :=
			engine.Request(methodSelect.Selected, urlEntry.Text)

		if err != nil {
			responseBox.SetText(err.Error())
			return
		}

		var pretty any
		if json.Unmarshal([]byte(body), &pretty) == nil {
			b, _ := json.MarshalIndent(pretty, "", "  ")
			responseBox.SetText(string(b))
		} else {
			responseBox.SetText(body)
		}

		w.SetTitle("LazyFyne Brain++ — " + status)

		rawLinks = newRaw
		smart = newSmart

		rawList.Refresh()
		smartList.Refresh()
	})

	left := container.NewVBox(
		widget.NewLabel("URL"),
		urlEntry,
		methodSelect,
		sendBtn,
		widget.NewSeparator(),
		widget.NewLabel("Detected Endpoints"),
		container.NewVScroll(smartList),
		widget.NewSeparator(),
		widget.NewLabel("Raw Links"),
		container.NewVScroll(rawList),
	)

	right := container.NewVBox(
		widget.NewLabel("Response"),
		container.NewVScroll(responseBox),
	)

	split := container.NewHSplit(left, right)
	split.Offset = 0.32

	w.SetContent(split)
	w.ShowAndRun()
}
