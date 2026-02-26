package entry

import "os"

// Normal should contain the logic that used to live in cmd/restless/main.go (old normal).
func Normal(args []string) error {
	// TODO: Paste/call your previous normal entry logic here.
	// Tip: If the old main used os.Args directly, replace with: os.Args = append([]string{os.Args[0]}, args...)
	_ = os.Args
	return nil
}
