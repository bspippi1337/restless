# Restless OpenAPI Guard + Auto Contract Sanitizer (merge patch)

This patch:
- Adds `restless openapi guard` + `restless openapi diff`
- Adds OpenAPI auto-discovery + memory+disk cache
- Adds a universal contract hook: `openapi.MaybeValidateResponse(...)`
- Adds a Termux-safe installer script

## How to enable sanitizing for *all* commands (probe/smart/raw)

Find the single place where HTTP responses are finalized (status, headers, body known),
typically in `internal/core/httpx` or `internal/core/engine`.

Add ONE call after you have:
- baseURL (scheme+host)
- method (GET/POST/...)
- pathTemplate (OpenAPI template if known; exact path works too)
- status
- content-type header
- response body bytes

```go
import (
  "context"
  openapi "github.com/bspippi1337/restless/internal/modules/openapi"
)

openapi.MaybeValidateResponse(
  context.Background(),
  baseURL,
  method,
  pathTemplate,
  status,
  contentType,
  bodyBytes,
)
```

Notes:
- Caching is automatic (memory + `~/.cache/restless/openapi/*.json`).
- If offline, disk cache is used.
- If no OpenAPI is discoverable, it silently no-ops.
