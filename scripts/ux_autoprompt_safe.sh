#!/usr/bin/env bash
set -euo pipefail

echo "==> Installing safe UX autoprompt + color"

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

func Color(code string, s string) string {
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", code, s)
}

func Status(code int) string {
	switch {
	case code >= 500:
		return Color("31", fmt.Sprintf("status: %d", code))
	case code >= 400:
		return Color("31", fmt.Sprintf("status: %d", code))
	case code >= 300:
		return Color("33", fmt.Sprintf("status: %d", code))
	default:
		return Color("32", fmt.Sprintf("status: %d", code))
	}
}
EOT

# Autoprompt helper
FILE="cmd/restless-v2/openapi_cli.go"

if ! grep -q "promptMissingPathParams" "$FILE"; then
cat >> "$FILE" <<'EOT'

func promptMissingPathParams(path string, params map[string]string) (map[string]string, error) {
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
		if _, ok := params[key]; !ok {
			missing = append(missing, key)
		}
		s = s[i+j+1:]
	}

	if len(missing) == 0 {
		return params, nil
	}

	if !term.IsTTY() {
		return nil, fmt.Errorf("missing path params: %v", missing)
	}

	reader := bufio.NewReader(os.Stdin)
	for _, k := range missing {
		fmt.Printf("Enter %s: ", k)
		v, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		params[k] = strings.TrimSpace(v)
	}
	return params, nil
}
EOT
fi

# Ensure imports exist
if ! grep -q 'internal/ui/term' "$FILE"; then
sed -i '/internal\/profile/a\
\t"github.com/bspippi1337/restless/internal/ui/term"' "$FILE"
fi

if ! grep -q '"bufio"' "$FILE"; then
sed -i '/import (/a\
\t"bufio"' "$FILE"
fi

# Replace ONLY status print line safely
sed -i 's/fmt.Printf("status: %d (dur=%dms)\\n", resp.StatusCode, resp.DurationMs)/fmt.Println(term.Status(resp.StatusCode), "(dur=", resp.DurationMs, "ms)")/' "$FILE"

gofmt -w .
go test ./...
go build -o restless-v2 ./cmd/restless-v2

echo "✅ Safe UX installed"
