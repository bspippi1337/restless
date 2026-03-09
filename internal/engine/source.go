package engine

func SourceScore(source string) string {
	switch source {
	case "probe":
		return "medium"
	case "crawler":
		return "medium"
	case "template":
		return "high"
	case "inferred":
		return "high"
	default:
		return "low"
	}
}
