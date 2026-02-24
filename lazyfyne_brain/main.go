package main

import (
	"encoding/json"
	"fmt"
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
	Client *http.Client
	LastJSON any
	LastURL string
}

func NewEngine() *Engine {
	return &Engine{
		Client: &http.Client{Timeout: 25 * time.Second},
	}
}

func (e *Engine) Request(method, target string) (string, string, []string, error) {
	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		return "", "", nil, err
	}

	req.Header.Set("Accept", "application/json, */*")

	resp, err := e.Client.Do(req)
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	body := string(bodyBytes)

	var parsed any
	if json.Unmarshal(bodyBytes, &parsed) == nil {
		e.LastJSON = parsed
		e.LastURL = target
	} else {
		e.LastJSON = nil
	}

	hints := buildHints(target, body)

	return resp.Status, body, hints, nil
}

func buildHints(currentURL, body string) []string {
	var hints []string

	reURL := regexp.MustCompile(`https?://[^\s"'<>]+`)
	for _, m := range reURL.FindAllString(body, 20) {
		hints = appendUnique(hints, m)
	}

	u, err := url.Parse(currentURL)
	if err == nil {
		parent := *u
		parent.Path = "/"
		hints = appendUnique(hints, parent.String())
	}

	return hints
}

func appendUnique(list []string, val string) []string {
	for _, v := range list {
		if v == val {
			return list
		}
	}
	return append(list, val)
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("LazyFyne Brain")
	w.Resize(fyne.NewSize(1300, 800))

	engine := NewEngine()

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://api.example.com")

	methodSelect := widget.NewSelect([]string{"GET", "POST", "PUT", "DELETE"}, nil)
	methodSelect.SetSelected("GET")

	responseBox := widget.NewMultiLineEntry()
	responseBox.Wrapping = fyne.TextWrapWord

	var hints []string
	var keyHints []string

	hintsList := widget.NewList(
		func() int { return len(hints) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(hints[i])
		},
	)

	hintsList.OnSelected = func(id widget.ListItemID) {
		urlEntry.SetText(hints[id])
	}

	keyList := widget.NewList(
		func() int { return len(keyHints) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(keyHints[i])
		},
	)

	keyList.OnSelected = func(id widget.ListItemID) {
		key := keyHints[id]
		urlEntry.SetText(generateFromKey(engine.LastURL, key))
		methodSelect.SetSelected("GET")
	}

	sendBtn := widget.NewButton("Send", func() {
		status, body, newHints, err := engine.Request(methodSelect.Selected, urlEntry.Text)
		if err != nil {
			responseBox.SetText(err.Error())
			return
		}

		var pretty any
		if json.Unmarshal([]byte(body), &pretty) == nil {
			b, _ := json.MarshalIndent(pretty, "", "  ")
			responseBox.SetText(string(b))
			keyHints = extractKeys(pretty)
			keyList.Refresh()
		} else {
			responseBox.SetText(body)
			keyHints = nil
			keyList.Refresh()
		}

		w.SetTitle("LazyFyne Brain — " + status)
		hints = newHints
		hintsList.Refresh()
	})

	left := container.NewVBox(
		widget.NewLabel("URL"),
		urlEntry,
		methodSelect,
		sendBtn,
		widget.NewSeparator(),
		widget.NewLabel("Next Endpoints"),
		container.NewVScroll(hintsList),
		widget.NewSeparator(),
		widget.NewLabel("JSON Key Navigation"),
		container.NewVScroll(keyList),
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

func extractKeys(v any) []string {
	var keys []string

	switch t := v.(type) {
	case map[string]any:
		for k := range t {
			keys = append(keys, k)
		}
	case []any:
		if len(t) > 0 {
			return extractKeys(t[0])
		}
	}
	return keys
}

func generateFromKey(baseURL, key string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	keyLower := strings.ToLower(key)

	if keyLower == "id" || strings.HasSuffix(keyLower, "_id") {
		return u.String() + "/1"
	}

	q := u.Query()
	q.Set(key, "true")
	u.RawQuery = q.Encode()

	return u.String()
}
