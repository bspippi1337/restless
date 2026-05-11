package table

import (
	"fmt"
	"strings"

	"github.com/bspippi1337/restless/internal/ui/viewport"
)

type Row struct {
	Left  string
	Right string
	Extra string
}

func Render(rows []Row) string {
	var b strings.Builder

	width := viewport.Width()
	left := 28

	if width < 72 {
		left = 22
	}

	if width >= 100 {
		left = 36
	}

	for _, row := range rows {
		line := fmt.Sprintf(
			"  %-*s %s",
			left,
			dotted(fit(row.Left, left), left),
			row.Right,
		)

		if row.Extra != "" {
			line += "  " + row.Extra
		}

		b.WriteString(strings.TrimRight(line, " "))
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
	return s[:width-1] + "…"
}
