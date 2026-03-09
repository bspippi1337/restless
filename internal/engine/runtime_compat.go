package engine

import (
	"fmt"
)

func NormalizeTarget(t string) string {
	return normalizeTarget(t)
}

func Step(i int, total int, msg string) {
	fmt.Printf("[%d/%d] %s\n", i, total, msg)
}
