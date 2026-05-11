package ui

import (
	"fmt"
	"strings"
)

func Section(title string) string {
	return fmt.Sprintf(
		"%s\n%s\n",
		title,
		strings.Repeat("-", len(title)),
	)
}

func KV(k, v string) string {
	return fmt.Sprintf(
		"%-12s %s\n",
		k,
		v,
	)
}

func Spacer() string {
	return "\n"
}

func Bullet(v string) string {
	return fmt.Sprintf(
		"  - %s\n",
		v,
	)
}

func Row(a string, b interface{}, c string) string {
	return fmt.Sprintf(
		"  %-20s %-3v %s\n",
		a,
		b,
		c,
	)
}
