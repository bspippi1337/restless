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
	Kind        string     `json:"kind"`      // "restless.snapshot.v1"
	CreatedAt   string     `json:"createdAt"` // RFC3339
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
		SpecPath:  specPath,
		Checked:   rep.Checked,
		Failed:    rep.Failed,
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
