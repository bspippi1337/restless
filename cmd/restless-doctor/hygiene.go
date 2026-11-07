package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// auditRoot enforces the repository layout without modifying the working tree.
// Conventional community documentation stays at root; old technical notes go to
// docs/ or archive/historical/ rather than accumulating in the project root.
func auditRoot(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	allowedDocs := map[string]bool{
		"README.md":          true,
		"CHANGELOG.md":       true,
		"CONTRIBUTING.md":    true,
		"CODE_OF_CONDUCT.md": true,
		"SECURITY.md":        true,
	}
	allowedScripts := map[string]bool{
		"install.sh": true, // legacy public entrypoint, separately audited
	}
	var unexpected []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		switch {
		case strings.HasSuffix(lower, ".md") && !allowedDocs[name]:
			unexpected = append(unexpected, name)
		case strings.HasSuffix(lower, ".sh") && !allowedScripts[name]:
			unexpected = append(unexpected, name)
		case name == "SHA" || name == "main":
			unexpected = append(unexpected, name)
		}
	}
	if len(unexpected) > 0 {
		sort.Strings(unexpected)
		return fmt.Errorf("root hygiene violation (%s): relocate historical material to %s", strings.Join(unexpected, ", "), filepath.Join("archive", "historical"))
	}
	return nil
}
