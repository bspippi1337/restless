package viewport

import "os"

func Width() int {
	if v := os.Getenv("RESTLESS_WIDTH"); v != "" {
		if n := atoi(v); n >= 48 {
			if n > 140 {
				return 140
			}
			return n
		}
	}

	return 72
}

func Compact() bool {
	return Width() < 72
}

func Wide() bool {
	return Width() >= 100
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = (n * 10) + int(r-'0')
	}
	return n
}
