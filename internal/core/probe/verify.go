package probe

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func Verify(ctx context.Context, method, fullURL string) (ok bool, status string, hint string) {
	mm := strings.ToUpper(strings.TrimSpace(method))
	if mm == "" { mm = "GET" }
	client := &http.Client{Timeout: 8 * time.Second}

	try := []string{mm}
	if mm != "HEAD" { try = append(try, "HEAD") }
	if mm != "OPTIONS" { try = append(try, "OPTIONS") }

	for _, m := range try {
		req, _ := http.NewRequestWithContext(ctx, m, fullURL, nil)
		req.Header.Set("User-Agent", "restless-probe/0.2")
		res, err := client.Do(req)
		if err != nil { continue }
		res.Body.Close()
		status = res.Status
		if res.StatusCode >= 500 { continue }
		if res.StatusCode == 401 || res.StatusCode == 403 {
			return true, status, "auth required"
		}
		return true, status, ""
	}
	return false, status, hint
}
