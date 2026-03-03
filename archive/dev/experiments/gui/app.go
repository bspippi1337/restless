package gui

import (
	"context"
	"encoding/json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func Run() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("Restless GUI")
	w.Resize(fyne.NewSize(1280, 820))

	bridge := NewExecBridge()

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://api.example.com")

	method := widget.NewSelect([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, nil)
	method.SetSelected("GET")

	status := widget.NewLabel("HTTP -")

	output := widget.NewMultiLineEntry()
	output.Disable()
	output.Wrapping = fyne.TextWrapWord

	var smart []string
	var links []string

	smartList := widget.NewList(
		func() int { return len(smart) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(smart[i])
		},
	)

	linkList := widget.NewList(
		func() int { return len(links) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(links[i])
		},
	)

	send := widget.NewButton("Send", func() {
		req := Request{
			Method: method.Selected,
			URL:    urlEntry.Text,
		}

		go func() {
			res, err := bridge.Do(context.Background(), req)

			fyne.Do(func() {
				if err != nil {
					status.SetText("ERROR")
					output.SetText(err.Error())
					return
				}

				status.SetText(res.StatusText)

				text := res.Stdout
				var j any
				if json.Unmarshal([]byte(text), &j) == nil {
					b, _ := json.MarshalIndent(j, "", "  ")
					text = string(b)
				}
				output.SetText(text)

				s, l := BuildHints(req.URL, res.Stdout)
				smart = s
				links = l
				smartList.Refresh()
				linkList.Refresh()
			})
		}()
	})

	left := container.NewVBox(
		widget.NewLabel("URL"),
		urlEntry,
		method,
		status,
		send,
		widget.NewSeparator(),
		widget.NewLabel("Detected Endpoints"),
		container.NewVScroll(smartList),
		widget.NewSeparator(),
		widget.NewLabel("Raw Links"),
		container.NewVScroll(linkList),
	)

	right := container.NewVBox(
		widget.NewLabel("Response"),
		container.NewVScroll(output),
	)

	split := container.NewHSplit(left, right)
	split.Offset = 0.34

	w.SetContent(split)
	w.ShowAndRun()
}
