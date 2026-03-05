package council

import "time"

type Finding struct {
	Engine     string
	Kind       string
	Target     string
	Evidence   string
	Confidence float64
	Timestamp  time.Time
}
