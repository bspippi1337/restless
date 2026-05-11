package restlesscore

import (
	"fmt"
	"strings"
)

func renderSection(b *strings.Builder, title string, rows []string) {
	if len(rows) == 0 {
		return
	}

	fmt.Fprintf(b, "%s\n", title)
	fmt.Fprintf(b, "%s\n", strings.Repeat("-", len(title)))

	for _, row := range rows {
		fmt.Fprintf(b, "  %s\n", row)
	}

	fmt.Fprintln(b)
}
