// RESTLESS 42 Desktop API Atlas
// A true native (non-WebView) Fyne workstation for Windows, macOS and Linux.
// Every network request is explicit, scope-bound and performed by Botanical.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/bspippi1337/restless/internal/botanical"
	"github.com/bspippi1337/restless/internal/library"
)

type atlasTheme struct{ fyne.Theme }

func (t atlasTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 183, G: 247, B: 104, A: 255}
	case theme.ColorNameBackground:
		return color.NRGBA{R: 13, G: 19, B: 29, A: 255}
	}
	return t.Theme.Color(n, v)
}

type screen struct {
	win        fyne.Window
	input      *widget.Entry
	output     *widget.Entry
	pages      *widget.Select
	depth      *widget.Select
	consent    *widget.Check
	status     *widget.Label
	counter    *widget.Label
	details    *widget.Entry
	logs       *widget.Entry
	search     *widget.Entry
	tree       *widget.Tree
	nodes      map[string][]string
	nodeLabels map[string]string
	operations map[string]botanical.Operation
	current    library.Summary
	inProgress bool
}

func main() {
	a := app.NewWithID("org.restless42.atlas.desktop")
	a.Settings().SetTheme(atlasTheme{Theme: theme.DarkTheme()})
	w := a.NewWindow("RESTLESS 42 · Native API Atlas")
	w.Resize(fyne.NewSize(1330, 820))
	s := &screen{win: w, nodes: map[string][]string{}, nodeLabels: map[string]string{}, operations: map[string]botanical.Operation{}}
	s.initialize()
	w.ShowAndRun()
}
func (s *screen) initialize() {
	s.input = widget.NewEntry()
	s.input.SetPlaceHolder("docs.example.org  or  https://docs.example.org/openapi.json")
	s.output = widget.NewEntry()
	s.output.SetText(".restless/libraries")
	s.pages = widget.NewSelect([]string{"1", "5", "12", "25"}, nil)
	s.pages.SetSelected("12")
	s.depth = widget.NewSelect([]string{"0", "1", "2", "3"}, nil)
	s.depth.SetSelected("2")
	s.consent = widget.NewCheck("I am authorized to retrieve documentation from this exact hostname", nil)
	s.status = widget.NewLabel("Ready. Preview a hostname before enabling network retrieval.")
	s.counter = widget.NewLabel("0 documents  ·  0 documented operations")
	s.details = widget.NewMultiLineEntry()
	s.details.SetMinRowsVisible(14)
	s.details.Disable()
	s.logs = widget.NewMultiLineEntry()
	s.logs.SetMinRowsVisible(7)
	s.logs.Disable()
	s.search = widget.NewEntry()
	s.search.SetPlaceHolder("Search a documented path or operation name")
	s.search.OnSubmitted = func(_ string) { s.searchFirst() }
	s.tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID { return s.nodes[id] },
		func(id widget.TreeNodeID) bool { return len(s.nodes[id]) > 0 },
		func(branch bool) fyne.CanvasObject { return widget.NewLabel("resource") },
		func(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			label := s.nodeLabels[id]
			if label == "" {
				label = id
			}
			obj.(*widget.Label).SetText(label)
		},
	)
	s.tree.OnSelected = func(id widget.TreeNodeID) {
		if op, ok := s.operations[id]; ok {
			s.selectOperation(op)
		} else if id != "" {
			s.details.SetText("RESOURCE\n" + id + "\n\nExpand this node to inspect documented HTTP methods.\n\nNo method is executed when you select it.")
		}
	}
	// Logo and consistent signal language, without embedding a website.
	brand := canvas.NewText("R.", color.NRGBA{R: 183, G: 247, B: 104, A: 255})
	brand.TextSize = 37
	brand.TextStyle = fyne.TextStyle{Bold: true}
	title := widget.NewRichTextFromMarkdown("# RESTLESS 42 / API ATLAS\n**One domain. Source-backed mapping. Zero fuzzing.**")
	header := container.NewHBox(brand, widget.NewSeparator(), title)
	preview := widget.NewButtonWithIcon("Preview scope", theme.SearchIcon(), s.preview)
	build := widget.NewButtonWithIcon("Build library", theme.DocumentCreateIcon(), s.build)
	importButton := widget.NewButtonWithIcon("Import library", theme.FolderOpenIcon(), s.openLibrary)
	exportButton := widget.NewButtonWithIcon("Save JSON", theme.DocumentSaveIcon(), s.exportLibrary)
	form := container.NewVBox(
		widget.NewLabelWithStyle("01 / TARGET AND SCOPE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Exact public domain or HTTPS docs URL", s.input),
			widget.NewFormItem("Output directory", s.output),
			widget.NewFormItem("Max documents", s.pages),
			widget.NewFormItem("Max link depth", s.depth),
		),
		s.consent,
		container.NewHBox(preview, build, importButton, exportButton),
		s.status, s.counter,
	)
	resourceHeader := container.NewVBox(
		widget.NewLabelWithStyle("02 / DOCUMENTED RESOURCE MAP", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		s.search,
		widget.NewButtonWithIcon("Find operation", theme.SearchIcon(), s.searchFirst),
	)
	left := container.NewBorder(resourceHeader, nil, nil, nil, s.tree)
	detailTitle := widget.NewLabelWithStyle("03 / SWAGGER-STYLE INSPECTOR + REQUEST EXAMPLES", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	detailPane := container.NewBorder(container.NewVBox(detailTitle, widget.NewLabel("All request examples are preview-only; no API operation is executed.")), nil, nil, nil, container.NewScroll(s.details))
	logPanel := container.NewBorder(widget.NewLabelWithStyle("ARCHIVE EVIDENCE / PROGRESS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, container.NewScroll(s.logs))
	right := container.NewVSplit(detailPane, logPanel)
	right.Offset = 0.72
	views := container.NewHSplit(left, right)
	views.Offset = 0.44
	s.win.SetContent(container.NewBorder(
		container.NewVBox(header, widget.NewSeparator(), form, widget.NewSeparator()),
		widget.NewLabel("Local library files: index.jsonl · sha256/ · catalog.json · library.json   |   Bot reads only permitted documentation URLs."),
		nil, nil, views))
}
func (s *screen) opts() library.Options {
	var n, d int
	fmt.Sscanf(s.pages.Selected, "%d", &n)
	fmt.Sscanf(s.depth.Selected, "%d", &d)
	return library.Options{
		Input: strings.TrimSpace(s.input.Text), OutputDir: strings.TrimSpace(s.output.Text),
		MaxPages: n, MaxDepth: d, Delay: 2 * time.Second, Execute: false, Authorized: false,
	}
}
func (s *screen) log(message string) {
	s.logs.SetText(s.logs.Text + time.Now().Format("15:04:05") + " " + message + "\n")
}
func (s *screen) preview() {
	out, err := library.Preview(s.opts())
	if err != nil {
		s.status.SetText("Scope error: " + err.Error())
		return
	}
	s.status.SetText("DRY RUN · " + out.Target.Host + " · No requests made")
	s.details.SetText(
		"SCOPE PREVIEW\n\nHost: " + out.Target.Host +
			"\nSeed: " + out.Target.SeedURL +
			"\nArchive: " + out.OutputDir +
			fmt.Sprintf("\nDocuments: max %d\nDepth: max %d\nDelay: minimum %ds\n", out.MaxPages, out.MaxDepth, out.DelaySeconds) +
			"\nUse Build library only with permission.\nNo domain guessing, no endpoint guessing.")
	s.log("Preview only: " + out.Target.SeedURL)
}
func (s *screen) build() {
	if s.inProgress {
		return
	}
	options := s.opts()
	if !s.consent.Checked {
		dialog.ShowInformation("Authorization required", "Confirm explicit permission and target scope before starting a real documentation crawl.", s.win)
		return
	}
	_, err := library.Preview(options)
	if err != nil {
		dialog.ShowError(err, s.win)
		return
	}
	options.Execute = true
	options.Authorized = true
	s.inProgress = true
	s.status.SetText("Collecting documented sources... respecting robots and the page budget.")
	s.log("Started authorized documentation crawl.")
	options.OnDocument = func(record botanical.Record) {
		fyne.Do(func() {
			s.log(fmt.Sprintf("ARCHIVED %s  SHA256 %s (%d bytes)", record.URL, record.SHA256, record.Bytes))
		})
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		result, err := library.Build(ctx, options)
		fyne.Do(func() {
			s.inProgress = false
			if err != nil {
				s.status.SetText("Crawl stopped: " + err.Error())
				s.log("ERROR " + err.Error())
				dialog.ShowError(err, s.win)
				return
			}
			s.current = result
			s.populate(result)
			s.log(fmt.Sprintf("Complete: %d documents / %d declared methods", result.DocumentCount, result.OperationCount))
			s.status.SetText("LIBRARY READY · Only declared API operations indexed.")
		})
	}()
}
func (s *screen) openLibrary() {
	dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, s.win)
			return
		}
		if r == nil {
			return
		}
		defer r.Close()
		data, err := io.ReadAll(io.LimitReader(r, 2<<20+1))
		if err != nil {
			dialog.ShowError(err, s.win)
			return
		}
		if len(data) > 2<<20 {
			dialog.ShowError(fmt.Errorf("library JSON exceeds 2 MB"), s.win)
			return
		}
		var result library.Summary
		if err = json.Unmarshal(data, &result); err != nil {
			dialog.ShowError(err, s.win)
			return
		}
		if result.SchemaVersion != 1 || result.Target.Host == "" {
			dialog.ShowError(fmt.Errorf("expected a RESTLESS library.json (schema_version 1)"), s.win)
			return
		}
		s.current = result
		s.populate(result)
		s.status.SetText("Imported library from local JSON file, no network requests.")
		s.log("Imported " + r.URI().Name() + " offline")
	}, s.win)
}
func (s *screen) exportLibrary() {
	if s.current.SchemaVersion != 1 {
		dialog.ShowInformation("No library", "Build or import a RESTLESS library before exporting.", s.win)
		return
	}
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, s.win)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		if err := enc.Encode(s.current); err != nil {
			dialog.ShowError(err, s.win)
			return
		}
		s.log("Exported provenance-backed library JSON")
	}, s.win)
}
func (s *screen) populate(summary library.Summary) {
	s.nodes = map[string][]string{}
	s.nodeLabels = map[string]string{}
	s.operations = map[string]botanical.Operation{}
	s.nodes[""] = []string{"/"}
	s.nodeLabels["/"] = summary.Target.Host
	seen := map[string]bool{"/": true}
	ops := append([]botanical.Operation(nil), summary.Operations...)
	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Path != ops[j].Path {
			return ops[i].Path < ops[j].Path
		}
		return ops[i].Method < ops[j].Method
	})
	for index, op := range ops {
		parent := "/"
		prefix := ""
		for _, segment := range strings.Split(strings.Trim(op.Path, "/"), "/") {
			if segment == "" {
				continue
			}
			prefix += "/" + segment
			if !seen[prefix] {
				s.nodes[parent] = append(s.nodes[parent], prefix)
				s.nodeLabels[prefix] = segment
				seen[prefix] = true
			}
			parent = prefix
		}
		id := fmt.Sprintf("op:%d:%s:%s", index, op.Method, op.Path)
		s.nodes[parent] = append(s.nodes[parent], id)
		s.nodeLabels[id] = op.Method + "  " + op.OperationID + "  " + op.Summary
		s.operations[id] = op
	}
	s.tree.Refresh()
	s.tree.OpenBranch("/")
	s.counter.SetText(fmt.Sprintf("%d evidence documents  ·  %d documented API operations", summary.DocumentCount, summary.OperationCount))
	s.details.SetText(fmt.Sprintf("API LIBRARY\nHost: %s\nSources: %d\nOperations: %d\n\nExplore the tree to view inputs, authentication declarations and source SHA-256 values.\n\nNo API methods are executed.", summary.Target.Host, summary.DocumentCount, summary.OperationCount))
}
func (s *screen) selectOperation(op botanical.Operation) {
	params := ""
	for _, p := range op.Parameters {
		typ := p.Schema.Type
		if typ == "" {
			typ = "string"
		}
		required := "optional"
		if p.Required {
			required = "required"
		}
		params += fmt.Sprintf("• %s (%s, %s, %s)\n  %s\n", p.Name, p.In, typ, required, p.Description)
	}
	if params == "" {
		params = "None declared in this source.\n"
	}
	host := s.current.Target.Host
	if host == "" {
		host = "YOUR-AUTHORIZED-HOST"
	}
	// The preview is a DOCUMENT, not an HTTP operation.
	endpoint := "https://" + host + op.Path
	example := "curl --request " + op.Method + " --url '" + strings.ReplaceAll(endpoint, "'", "'\\''") + "' --header 'Accept: application/json'"
	 requestNotes := "No request body declared.\n"
 if op.Request != nil {
  b,err:=json.MarshalIndent(op.Request,"","  ")
  if err==nil {requestNotes=string(b)+"\n"}
 }
 responseNotes := "No response types declared.\n"
 if len(op.Responses)>0 {
  b,err:=json.MarshalIndent(op.Responses,"","  ")
  if err==nil {responseNotes=string(b)+"\n"}
 }
detail := fmt.Sprintf(
		"DOCUMENTED API OPERATION\n\n%s %s\n%s\n\nOperation ID: %s\n\n"+
			"INPUT PARAMETERS\n%s\nDECLARED SECURITY\n%v\n\nUNRESOLVED REFERENCES: %t\n\n"+
			"DOCUMENTED REQUEST BODY\n%s\nDOCUMENTED RESPONSES\n%s\n"+
			"REQUEST EXAMPLE / PREVIEW ONLY\n%s\n\n"+
			"SOURCE EVIDENCE\nURL: %s\nSHA256: %s\n\n"+
			"IMPORTANT: This example is not executed. Replace placeholders, verify authorization, declared servers and effects before using any API client.",
		op.Method, op.Path, op.Summary, op.OperationID, params, op.Security, op.UnresolvedRefs,
		requestNotes, responseNotes, example, op.SourceURL, op.SourceSHA256)
	s.details.SetText(detail)
}
func (s *screen) searchFirst() {
	query := strings.ToLower(strings.TrimSpace(s.search.Text))
	if query == "" {
		return
	}
	keys := make([]string, 0, len(s.operations))
	for key := range s.operations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		op := s.operations[key]
		if strings.Contains(strings.ToLower(op.Path+" "+op.Method+" "+op.OperationID+" "+op.Summary), query) {
			s.tree.Select(key)
			s.selectOperation(op)
			s.status.SetText("Selected documented operation " + op.Method + " " + op.Path)
			return
		}
	}
	s.status.SetText("No documented operation matches: " + query)
}
