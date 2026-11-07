package botanical

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ContractChange reports an operation-level difference without network requests.
// References keep the original source URL and content hashes for auditability.
type ContractChange struct {
	Key    string     `json:"key"`
	Before *Operation `json:"before,omitempty"`
	After  *Operation `json:"after,omitempty"`
}

type ContractDiff struct {
	Mode            string           `json:"mode"`
	NetworkRequests int              `json:"network_requests"`
	Added           []ContractChange `json:"added"`
	Removed         []ContractChange `json:"removed"`
	Changed         []ContractChange `json:"changed"`
	Unchanged       int              `json:"unchanged"`
	Notes           []string         `json:"notes"`
}

// CompareCatalogs compares API contracts, not raw document-byte hashes.
// A cosmetic documentation change with the same operation fields is unchanged.
func CompareCatalogs(before, after Catalog) (ContractDiff, error) {
	report := ContractDiff{
		Mode:            "OFFLINE_CONTRACT_DIFF",
		NetworkRequests: 0,
		Notes: []string{
			"Operation identity = source URL + HTTP method + documented path.",
			"Changes compare documented operation fields and ignore archive SHA-256.",
			"Renamed or relocated source URLs appear as removals and additions.",
			"No HTTP requests, live probes, or server compatibility claims.",
		},
	}
	if before.Version != catalogVersion || after.Version != catalogVersion {
		return report, fmt.Errorf("catalog version mismatch: expected %d", catalogVersion)
	}
	old, err := operationMap(before.Operations)
	if err != nil {
		return report, err
	}
	newOps, err := operationMap(after.Operations)
	if err != nil {
		return report, err
	}
	var keys []string
	seen := map[string]bool{}
	for key := range old {
		keys = append(keys, key)
		seen[key] = true
	}
	for key := range newOps {
		if !seen[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		a, aOK := old[key]
		b, bOK := newOps[key]
		switch {
		case !aOK:
			copy := b
			report.Added = append(report.Added, ContractChange{Key: key, After: &copy})
		case !bOK:
			copy := a
			report.Removed = append(report.Removed, ContractChange{Key: key, Before: &copy})
		default:
			aCanonical, err := canonicalOperation(a)
			if err != nil {
				return report, err
			}
			bCanonical, err := canonicalOperation(b)
			if err != nil {
				return report, err
			}
			if aCanonical == bCanonical {
				report.Unchanged++
			} else {
				previous, current := a, b
				report.Changed = append(report.Changed, ContractChange{
					Key: key, Before: &previous, After: &current,
				})
			}
		}
	}
	return report, nil
}

func operationMap(operations []Operation) (map[string]Operation, error) {
	byKey := make(map[string]Operation, len(operations))
	for _, op := range operations {
		if op.SourceURL == "" || op.Method == "" || op.Path == "" {
			return nil, fmt.Errorf("operation missing identity or provenance")
		}
		key := op.SourceURL + " " + op.Method + " " + op.Path
		if _, exists := byKey[key]; exists {
			return nil, fmt.Errorf("duplicate operation key: %s", key)
		}
		byKey[key] = op
	}
	return byKey, nil
}

func canonicalOperation(op Operation) (string, error) {
	op.SourceSHA256 = "" // Hash changes alone are not semantic API changes.
	b, err := json.Marshal(op)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
