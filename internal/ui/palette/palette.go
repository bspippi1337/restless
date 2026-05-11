package palette

const (
	Reset  = "\x1b[0m"
	Bold   = "\x1b[1m"
	Dim    = "\x1b[2m"
	Cyan   = "\x1b[36m"
	Blue   = "\x1b[34m"
	Green  = "\x1b[32m"
	Yellow = "\x1b[33m"
	Red    = "\x1b[31m"
	Gray   = "\x1b[90m"
)

func Header(s string) string {
	return Bold + Cyan + s + Reset
}

func Section(s string) string {
	return Bold + Blue + s + Reset
}

func Live(s string) string {
	return Green + s + Reset
}

func Warn(s string) string {
	return Yellow + s + Reset
}

func Danger(s string) string {
	return Red + s + Reset
}

func Muted(s string) string {
	return Gray + s + Reset
}
