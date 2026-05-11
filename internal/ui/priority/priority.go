package priority

const (
	Background = iota
	Midground
	Foreground
	Critical
)

type Weighted struct {
	Text   string
	Weight int
}

func Rank(status int) int {
	switch {
	case status >= 500:
		return Critical
	case status >= 400:
		return Foreground
	case status >= 200:
		return Midground
	default:
		return Background
	}
}

func Emphasis(weight int) string {
	switch weight {
	case Critical:
		return "critical"
	case Foreground:
		return "focus"
	case Midground:
		return "normal"
	default:
		return "muted"
	}
}
