package diff

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/bspippi1337/restless/internal/snapshot"
)

type ChangeKind string

const (
	Added   ChangeKind = "+"
	Removed ChangeKind = "-"
	Changed ChangeKind = "~"
)

type Change struct {
	Kind   ChangeKind `json:"kind"`
	Method string     `json:"method"`
	Path   string     `json:"path"`

	FromCode int `json:"fromCode,omitempty"`
	ToCode   int `json:"toCode,omitempty"`
}

type Report struct {
	Same        bool     `json:"same"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	FromFP      string   `json:"fromFingerprint"`
	ToFP        string   `json:"toFingerprint"`
	ChangeCount int      `json:"changeCount"`
	Changes     []Change `json:"changes"`
}

func key(m, p string) string {
	return strings.ToUpper(m) + " " + p
}

func Compare(a, b snapshot.Snapshot) Report {
	am := map[string]snapshot.Endpoint{}
	bm := map[string]snapshot.Endpoint{}

	for _, e := range a.Endpoints {
		am[key(e.Method, e.Path)] = e
	}
	for _, e := range b.Endpoints {
		bm[key(e.Method, e.Path)] = e
	}

	var changes []Change

	// removed / changed
	for k, ae := range am {
		be, ok := bm[k]
		if !ok {
			changes = append(changes, Change{Kind: Removed, Method: ae.Method, Path: ae.Path})
			continue
		}
		if ae.ActualCode != be.ActualCode || ae.ExpectedCodes != be.ExpectedCodes {
			changes = append(changes, Change{
				Kind:     Changed,
				Method:   ae.Method,
				Path:     ae.Path,
				FromCode: ae.ActualCode,
				ToCode:   be.ActualCode,
			})
		}
	}

	// added
	for k, be := range bm {
		if _, ok := am[k]; !ok {
			changes = append(changes, Change{Kind: Added, Method: be.Method, Path: be.Path})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			if changes[i].Method == changes[j].Method {
				return changes[i].Kind < changes[j].Kind
			}
			return changes[i].Method < changes[j].Method
		}
		return changes[i].Path < changes[j].Path
	})

	return Report{
		Same:        len(changes) == 0 && a.Fingerprint == b.Fingerprint,
		From:        a.BaseURL,
		To:          b.BaseURL,
		FromFP:      a.Fingerprint,
		ToFP:        b.Fingerprint,
		ChangeCount: len(changes),
		Changes:     changes,
	}
}

func PrintHuman(w io.Writer, r Report) {
	if r.Same {
		fmt.Fprintf(w, "✔ diff OK (no drift)\n")
		fmt.Fprintf(w, "  from: %s  fp=%s\n", r.From, r.FromFP)
		fmt.Fprintf(w, "  to:   %s  fp=%s\n", r.To, r.ToFP)
		return
	}
	fmt.Fprintf(w, "✖ diff DRIFT detected (%d)\n\n", r.ChangeCount)
	for _, c := range r.Changes {
		switch c.Kind {
		case Added:
			fmt.Fprintf(w, "+ %s %s\n", c.Method, c.Path)
		case Removed:
			fmt.Fprintf(w, "- %s %s\n", c.Method, c.Path)
		case Changed:
			fmt.Fprintf(w, "~ %s %s  (%d → %d)\n", c.Method, c.Path, c.FromCode, c.ToCode)
		default:
			fmt.Fprintf(w, "? %s %s\n", c.Method, c.Path)
		}
	}
}
