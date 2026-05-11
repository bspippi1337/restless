package fuzzwow

func enrichGitHub(r *Result) {
	r.Signals = appendUnique(
		r.Signals,
		"github edge shielding",
	)

	r.Signals = appendUnique(
		r.Signals,
		"authenticated traversal preferred",
	)

	if len(r.Blocked) >= 8 {
		r.Signals = appendUnique(
			r.Signals,
			"highly defended public edge",
		)
	}
}

func appendUnique(in []string, v string) []string {
	for _, x := range in {
		if x == v {
			return in
		}
	}

	return append(in, v)
}
