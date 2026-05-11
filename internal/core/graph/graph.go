package graph

import (
	"strings"

	"github.com/bspippi1337/restless/internal/core/state"
)

func Render() (string, error) {

	st, _, err := state.Load()
	if err != nil {
		return "", err
	}

	var b strings.Builder

	b.WriteString("API MAP\n")
	b.WriteString("=======\n")

	for _, r := range st.LastScan.Endpoints {

		b.WriteString("  ")
		b.WriteString(r.Method)
		b.WriteString(" ")
		b.WriteString(r.Path)
		b.WriteString("\n")

	}

	return b.String(), nil
}
