package layout

import (
	"fmt"
	"strings"
)

type Document struct {
	Title string
	Meta  [][2]string
	Parts []Section
}

type Section struct {
	Title string
	Rows  []Row
	Items []string
}

type Row struct {
	Name  string
	Value string
	Note  string
}

func Render(doc Document) string {
	var b strings.Builder
	width := 64

	b.WriteString(strings.ToUpper(doc.Title))
	b.WriteByte('\n')
	b.WriteString(strings.Repeat("=", width))
	b.WriteString("\n\n")

	for _, meta := range doc.Meta {
		b.WriteString(fmt.Sprintf("%-10s %s\n", strings.ToUpper(meta[0]), meta[1]))
	}

	if len(doc.Meta) > 0 {
		b.WriteByte('\n')
	}

	for _, part := range doc.Parts {
		if len(part.Rows) == 0 && len(part.Items) == 0 {
			continue
		}

		b.WriteString(strings.ToUpper(part.Title))
		b.WriteByte('\n')
		b.WriteString(strings.Repeat("-", width))
		b.WriteString("\n\n")

		for _, row := range part.Rows {
			name := fit(row.Name, 28)
			value := row.Value
			if row.Note != "" {
				value += "  " + row.Note
			}
			b.WriteString(fmt.Sprintf("  %-28s %s\n", dotted(name, 28), value))
		}

		for _, item := range part.Items {
			b.WriteString("  - ")
			b.WriteString(item)
			b.WriteByte('\n')
		}

		b.WriteByte('\n')
	}

	return b.String()
}

func dotted(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(".", width-len(s))
}

func fit(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 1 {
		return s[:width]
	}
	return s[:width-1] + "."
}
