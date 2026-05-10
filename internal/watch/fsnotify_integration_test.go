package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherSeesFilesystemMutation(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.txt")

	if err := os.WriteFile(file, []byte("restless\n"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	st, err := os.Stat(file)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	if st.Size() == 0 {
		t.Fatalf("expected non-empty file")
	}

	if err := os.WriteFile(file, []byte("restless elite\n"), 0644); err != nil {
		t.Fatalf("rewrite failed: %v", err)
	}

	time.Sleep(25 * time.Millisecond)

	st2, err := os.Stat(file)
	if err != nil {
		t.Fatalf("restat failed: %v", err)
	}

	if st2.ModTime().Before(st.ModTime()) {
		t.Fatalf("modtime regression detected")
	}
}
