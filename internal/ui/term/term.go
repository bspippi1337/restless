package term

import (
	"fmt"
	"os"
)

func IsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func Color(code string, s string) string {
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", code, s)
}

func Status(code int) string {
	switch {
	case code >= 500:
		return Color("31", fmt.Sprintf("status: %d", code))
	case code >= 400:
		return Color("31", fmt.Sprintf("status: %d", code))
	case code >= 300:
		return Color("33", fmt.Sprintf("status: %d", code))
	default:
		return Color("32", fmt.Sprintf("status: %d", code))
	}
}
