package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
)

type FindingKey struct {
	OpID     string `json:"op_id"`
	Kind     string `json:"kind"`
	JSONPath string `json:"json_path"`
	Message  string `json:"message"`
	Status   int    `json:"status"`
}

type FindingStat struct {
	Key       FindingKey `json:"key"`
	Count     int        `json:"count"`
	FirstSeen time.Time  `json:"first_seen"`
	LastSeen  time.Time  `json:"last_seen"`
	LastCDI   float64    `json:"last_cdi"`
	AvgCDI    float64    `json:"avg_cdi"`
}

type HostSnapshot struct {
	BaseURL     string                  `json:"base_url"`
	SpecRef     string                  `json:"spec_ref"`
	UpdatedAt   time.Time               `json:"updated_at"`
	TotalEvents int                     `json:"total_events"`
	Findings    map[string]*FindingStat `json:"findings"` // keyHash -> stats
}

func cacheDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".cache", "restless", "ai")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func snapshotPath(baseURL string) string {
	h := sha256.Sum256([]byte(baseURL))
	return filepath.Join(cacheDir(), hex.EncodeToString(h[:])+".json")
}

func Load(baseURL string) (*HostSnapshot, error) {
	p := snapshotPath(baseURL)
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var s HostSnapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Findings == nil {
		s.Findings = map[string]*FindingStat{}
	}
	return &s, nil
}

func Save(s *HostSnapshot) error {
	if s == nil {
		return nil
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(snapshotPath(s.BaseURL), b, 0644)
}

func UpdateFromGuard(baseURL, specRef string, findings []model.Finding, cdi float64) error {
	now := time.Now()

	s, _ := Load(baseURL)
	if s == nil {
		s = &HostSnapshot{
			BaseURL:  baseURL,
			SpecRef:  specRef,
			Findings: map[string]*FindingStat{},
		}
	}
	s.SpecRef = specRef
	s.UpdatedAt = now

	for _, f := range findings {
		key := FindingKey{
			OpID:     f.OpID,
			Kind:     string(f.Kind),
			JSONPath: f.JSONPath,
			Message:  f.Message,
			Status:   f.Status,
		}
		h := hashKey(key)
		st, ok := s.Findings[h]
		if !ok {
			st = &FindingStat{
				Key:       key,
				FirstSeen: now,
				LastSeen:  now,
				LastCDI:   cdi,
				AvgCDI:    cdi,
			}
			s.Findings[h] = st
		}
		st.Count++
		st.LastSeen = now
		st.LastCDI = cdi
		st.AvgCDI = ((st.AvgCDI * float64(st.Count-1)) + cdi) / float64(st.Count)
		s.TotalEvents++
	}

	return Save(s)
}

func hashKey(k FindingKey) string {
	b, _ := json.Marshal(k)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func TopFindings(s *HostSnapshot, minCount int) []*FindingStat {
	if s == nil {
		return nil
	}
	out := make([]*FindingStat, 0, len(s.Findings))
	for _, v := range s.Findings {
		if v.Count >= minCount {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].AvgCDI > out[j].AvgCDI
		}
		return out[i].Count > out[j].Count
	})
	return out
}
