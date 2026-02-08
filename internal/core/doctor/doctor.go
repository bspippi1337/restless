package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Run(root string, dryRun bool) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" { root = "." }
	root, _ = filepath.Abs(root)

	start := time.Now()
	actions := []string{}
	removed := 0

	for _, d := range []string{"bin", "dist", "build", "logs", ".fixall-logs"} {
		p := filepath.Join(root, d)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			actions = append(actions, fmt.Sprintf("rm -rf %s", p))
			if !dryRun {
				_ = os.RemoveAll(p)
			}
			removed++
		}
	}

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil { return nil }
		if d.IsDir() {
			bn := filepath.Base(path)
			if bn == ".git" || bn == "node_modules" { return filepath.SkipDir }
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".log") {
			actions = append(actions, fmt.Sprintf("rm %s", path))
			if !dryRun {
				_ = os.Remove(path)
			}
			removed++
		}
		return nil
	})

	rep := &strings.Builder{}
	fmt.Fprintf(rep, "==> Restless Doctor (v2)\nRoot: %s\n\n", root)
	if dryRun {
		fmt.Fprintln(rep, "[DRY RUN] No files were deleted.")
	}
	if len(actions) == 0 {
		fmt.Fprintln(rep, "[ OK ] Nothing to clean.")
	} else {
		fmt.Fprintln(rep, "Actions:")
		for _, a := range actions {
			fmt.Fprintf(rep, "  - %s\n", a)
		}
		fmt.Fprintf(rep, "\n[ OK ] Cleaned targets (%d actions).\n", removed)
	}
	fmt.Fprintf(rep, "Done in %s\n", time.Since(start).Truncate(10*time.Millisecond))
	fmt.Fprintln(rep, "Next:")
	fmt.Fprintln(rep, "  - make build")
	fmt.Fprintln(rep, "  - restless (TUI)")
	fmt.Fprintln(rep, "  - restless discover <domain> --json")
	return rep.String(), nil
}
