#!/usr/bin/env bash
set -euo pipefail

MAIN="cmd/restless/main.go"

[ -f go.mod ] || { echo "ERROR: run from repo root (missing go.mod)"; exit 1; }
[ -f "$MAIN" ] || { echo "ERROR: $MAIN not found"; exit 1; }

mkdir -p internal/snapshot internal/diff

# -----------------------------
# 1) SNAPSHOT FORMAT + FINGERPRINT
# -----------------------------
cat > internal/snapshot/snapshot.go <<'GO'
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/validate"
)

type Endpoint struct {
	Method        string `json:"method"`
	Path          string `json:"path"`
	ExpectedCodes string `json:"expectedCodes,omitempty"`
	ActualCode    int    `json:"actualCode"`
}

type Snapshot struct {
	Kind        string     `json:"kind"`        // "restless.snapshot.v1"
	CreatedAt   string     `json:"createdAt"`   // RFC3339
	BaseURL     string     `json:"baseUrl"`
	SpecPath    string     `json:"specPath"`
	Checked     int        `json:"checked"`
	Failed      int        `json:"failed"`
	Endpoints   []Endpoint `json:"endpoints"`   // sorted deterministically
	Fingerprint string     `json:"fingerprint"` // sha256 over normalized endpoints
}

func FromValidateReport(baseURL, specPath string, rep validate.Report) Snapshot {
	eps := make([]Endpoint, 0, rep.Checked)

	// Build map of failures first
	failMap := map[string]validate.Finding{}
	for _, f := range rep.Findings {
		key := strings.ToUpper(f.Method) + " " + f.Path
		failMap[key] = f
	}

	// We only have explicit per-endpoint info for failures in validate.Report today.
	// So we snapshot failures + include a stable fingerprint of "no drift" when OK.
	// Future enhancement: validate can be extended to return per-endpoint observed codes.
	for _, f := range rep.Findings {
		eps = append(eps, Endpoint{
			Method:        strings.ToUpper(f.Method),
			Path:          f.Path,
			ExpectedCodes: f.ExpectedCodes,
			ActualCode:    f.ActualCode,
		})
	}

	sort.Slice(eps, func(i, j int) bool {
		if eps[i].Path == eps[j].Path {
			return eps[i].Method < eps[j].Method
		}
		return eps[i].Path < eps[j].Path
	})

	s := Snapshot{
		Kind:      "restless.snapshot.v1",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		BaseURL:   baseURL,
		SpecPath: specPath,
		Checked:  rep.Checked,
		Failed:   rep.Failed,
		Endpoints: eps,
	}
	s.Fingerprint = Fingerprint(s)
	return s
}

func Fingerprint(s Snapshot) string {
	// Normalize only what matters for drift comparisons
	type nE struct {
		M string `json:"m"`
		P string `json:"p"`
		E string `json:"e,omitempty"`
		A int    `json:"a"`
	}
	tmp := make([]nE, 0, len(s.Endpoints))
	for _, e := range s.Endpoints {
		tmp = append(tmp, nE{
			M: strings.ToUpper(e.Method),
			P: e.Path,
			E: e.ExpectedCodes,
			A: e.ActualCode,
		})
	}
	sort.Slice(tmp, func(i, j int) bool {
		if tmp[i].P == tmp[j].P {
			return tmp[i].M < tmp[j].M
		}
		return tmp[i].P < tmp[j].P
	})
	b, _ := json.Marshal(tmp)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func WriteJSON(path string, s Snapshot) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func ReadJSON(path string) (Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return Snapshot{}, err
	}
	defer f.Close()
	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return Snapshot{}, err
	}
	if s.Kind == "" {
		return Snapshot{}, fmt.Errorf("invalid snapshot: missing kind")
	}
	return s, nil
}

func PrintHuman(w io.Writer, s Snapshot) {
	fmt.Fprintf(w, "snapshot %s\n", s.Kind)
	fmt.Fprintf(w, "  base: %s\n  spec: %s\n  checked: %d\n  drift: %d\n  fingerprint: %s\n",
		s.BaseURL, s.SpecPath, s.Checked, s.Failed, s.Fingerprint)
}
GO

# -----------------------------
# 2) DIFF ENGINE
# -----------------------------
cat > internal/diff/diff.go <<'GO'
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
GO

# -----------------------------
# 3) WIRE COMMANDS INTO main.go (switch router)
# -----------------------------

# Ensure imports we need
if ! grep -q 'internal/snapshot' "$MAIN"; then
  perl -0777 -i -pe 's|(github\.com/bspippi1337/restless/internal/modules/session"\n)|$1\t"github.com/bspippi1337/restless/internal/snapshot"\n\t"github.com/bspippi1337/restless/internal/diff"\n|s' "$MAIN"
fi

# Add cases snapshot + diff into switch
if ! grep -q 'case "snapshot"' "$MAIN"; then
  perl -0777 -i -pe 's/(case "profile":\s*handleProfile\(os\.Args\[2:\]\)\s*return\s*)/$1\n\t\tcase "snapshot":\n\t\t\thandleSnapshot(os.Args[2:])\n\t\t\treturn\n\t\tcase "diff":\n\t\t\thandleDiff(os.Args[2:])\n\t\t\treturn\n/s' "$MAIN"
fi

# Add handlers if not present
if ! grep -q 'func handleSnapshot' "$MAIN"; then
cat >> "$MAIN" <<'GO'

func handleSnapshot(args []string) {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)

	spec := fs.String("spec", "", "Path to OpenAPI spec")
	base := fs.String("base", "", "Base URL")
	out := fs.String("out", "snapshot.json", "Output file")
	timeout := fs.Int("timeout", 7, "Timeout in seconds")
	strict := fs.Bool("strict", false, "Strict mode (fail on unexpected status codes)")
	jsonOut := fs.Bool("json", false, "Print snapshot JSON to stdout")

	fs.Parse(args)

	if *spec == "" || *base == "" {
		fmt.Println("missing --spec or --base")
		fs.Usage()
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	rep, err := validate.Run(ctx, validate.Options{
		SpecPath:   *spec,
		BaseURL:    *base,
		Timeout:    time.Duration(*timeout) * time.Second,
		StrictLive: *strict,
	})
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}

	snap := snapshot.FromValidateReport(*base, *spec, rep)

	// Always write file (so CI artifacts exist)
	if err := snapshot.WriteJSON(*out, snap); err != nil {
		fmt.Println("ERROR: write snapshot:", err)
		os.Exit(2)
	}

	if *jsonOut {
		b, _ := os.ReadFile(*out)
		fmt.Print(string(b))
	} else {
		snapshot.PrintHuman(os.Stdout, snap)
		fmt.Printf("  wrote: %s\n", *out)
	}

	// Snapshot itself does not fail the build. Use `restless diff` or `restless validate` for gating.
}

func handleDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	aPath := fs.String("a", "", "Snapshot A path")
	bPath := fs.String("b", "", "Snapshot B path")

	// convenience: allow positional args: restless diff a.json b.json
	fs.Parse(args)
	rest := fs.Args()

	if *aPath == "" && len(rest) >= 1 {
		*aPath = rest[0]
	}
	if *bPath == "" && len(rest) >= 2 {
		*bPath = rest[1]
	}

	if *aPath == "" || *bPath == "" {
		fmt.Println("usage: restless diff A.snap.json B.snap.json")
		fs.Usage()
		os.Exit(2)
	}

	a, err := snapshot.ReadJSON(*aPath)
	if err != nil {
		fmt.Println("ERROR: read A:", err)
		os.Exit(2)
	}
	b, err := snapshot.ReadJSON(*bPath)
	if err != nil {
		fmt.Println("ERROR: read B:", err)
		os.Exit(2)
	}

	r := diff.Compare(a, b)
	diff.PrintHuman(os.Stdout, r)

	if !r.Same {
		os.Exit(1)
	}
}
GO
fi

# Ensure time import exists (main.go already uses time in your timeout patch; safe to ensure)
if ! grep -q '"time"' "$MAIN"; then
  perl -0777 -i -pe 's/import \(/import (\n\t"time"/;' "$MAIN"
fi

echo "==> Formatting..."
gofmt -w internal cmd >/dev/null

echo "==> Building..."
go build ./cmd/restless

echo "==> Committing..."
git add -A
git commit -m "feat: snapshot + diff (deterministic API drift forensics)" || true

echo
echo "✅ Installed commands:"
echo
echo "  restless snapshot --spec openapi.yaml --base https://api.example.com --out prod.snap.json"
echo "  restless snapshot --spec openapi.yaml --base https://staging.api.example.com --out staging.snap.json"
echo "  restless diff staging.snap.json prod.snap.json"
