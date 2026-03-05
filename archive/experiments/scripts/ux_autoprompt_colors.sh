#!/usr/bin/env bash
set -euo pipefail

echo "==> UX: autoprompt missing path params + colorful output"

# Add tiny term helper
mkdir -p internal/ui/term
cat > internal/ui/term/term.go <<'EOT'
package term

import (
	"fmt"
	"os"
)

func IsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func Colorize(enabled bool, code string, s string) string {
	if !enabled {
		return s
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", code, s)
}

// Some convenient styles
func Green(enabled bool, s string) string  { return Colorize(enabled, "32", s) }
func Yellow(enabled bool, s string) string { return Colorize(enabled, "33", s) }
func Red(enabled bool, s string) string    { return Colorize(enabled, "31", s) }
func Cyan(enabled bool, s string) string   { return Colorize(enabled, "36", s) }
func Dim(enabled bool, s string) string    { return Colorize(enabled, "2", s) }
func Bold(enabled bool, s string) string   { return Colorize(enabled, "1", s) }
EOT

# Patch openapi_cli.go run block: autoprompt + color line
FILE="cmd/restless-v2/openapi_cli.go"

# Ensure imports
if ! grep -q 'internal/ui/term' "$FILE"; then
  sed -i '/internal\/profile/a\
\t"github.com/bspippi1337/restless/internal/ui/term"' "$FILE"
fi

if ! grep -q 'bufio' "$FILE"; then
  sed -i '/import (/a\
\t"bufio"\n\t"strings"\n\t"os"' "$FILE"
fi

# Insert autoprompt helper in file (once)
if ! grep -q 'func promptMissingPathParams' "$FILE"; then
  cat >> "$FILE" <<'EOT'

func promptMissingPathParams(path string, params map[string]string) (map[string]string, error) {
	// Find {param} occurrences
	missing := []string{}
	s := path
	for {
		i := strings.Index(s, "{")
		if i == -1 {
			break
		}
		j := strings.Index(s[i:], "}")
		if j == -1 {
			break
		}
		key := s[i+1 : i+j]
		if _, ok := params[key]; !ok && key != "" {
			missing = append(missing, key)
		}
		s = s[i+j+1:]
	}

	if len(missing) == 0 {
		return params, nil
	}

	// If not interactive, fail loudly
	if !term.IsTTY() {
		return nil, fmt.Errorf("missing path params: %v (non-interactive, pass -p key=value)", missing)
	}

	in := bufio.NewReader(os.Stdin)
	for _, k := range missing {
		fmt.Printf("Enter %s: ", k)
		val, err := in.ReadString('\n')
		if err != nil {
			return nil, err
		}
		val = strings.TrimSpace(val)
		if val == "" {
			return nil, fmt.Errorf("empty value for path param: %s", k)
		}
		params[k] = val
	}
	return params, nil
}
EOT
fi

# Inject autoprompt call right before BuildRequest in the run case
# We look for the line that calls BuildRequest and insert above it.
perl -0777 -i -pe 's/\n\s*req, curl, err := openapi\.BuildRequest\(/\
\n    // Autoprompt missing path params (interactive when TTY)\n    ra.PathParams, err = promptMissingPathParams(ra.Path, ra.PathParams)\n    if err != nil {\n        fmt.Println(\"ERROR: params:\", err)\n        os.Exit(1)\n    }\n\n    req, curl, err := openapi.BuildRequest(/s' "$FILE"

# Make output colored in run case: replace status print (best effort)
perl -0777 -i -pe 's/fmt\.Printf\("status: %d \(dur=%dms\)\\n", resp\.StatusCode, resp\.DurationMs\)/\
{\
    color := term.IsTTY()\
    code := \"32\"\
    if resp.StatusCode >= 400 { code = \"31\" } else if resp.StatusCode >= 300 { code = \"33\" }\
    line := fmt.Sprintf(\"status: %d (dur=%dms)\", resp.StatusCode, resp.DurationMs)\
    fmt.Println(term.Bold(color, term.Colorize(color, code, line)))\
    fmt.Println(term.Dim(color, \"---\"))\
}/s' "$FILE"

gofmt -w internal/ui/term/term.go cmd/restless-v2/openapi_cli.go
go test ./...
go build -o restless-v2 ./cmd/restless-v2

echo "✅ UX done: autoprompt + colorful status line"
