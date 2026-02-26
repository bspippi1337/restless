# Restless ⚡

**curl sends requests. Restless helps you understand the API.**

Restless is a terminal-first API workbench written in Go.  
It adapts to the API surface:

- If there’s an OpenAPI spec, it becomes spec-driven.
- If there isn’t, it switches to heuristic discovery.

---

## Quick Demo

```bash
restless probe https://httpbin.org
restless list
restless run GET /headers
restless session
```

Want a 15-second terminal recording for README? See `docs/DEMO.md`.

---

## Installation

```bash
go install github.com/bspippi1337/restless/cmd/restless@latest
```

Or build locally:

```bash
git clone https://github.com/bspippi1337/restless
cd restless
go build ./cmd/restless
```

---

## Project Structure

```text
.
├── cmd/
│   └── restless/        # CLI entrypoint
├── internal/            # core engine + UI
├── docs/
└── README.md
```

Entry point: `cmd/restless/main.go`

---

## What works today

- `probe`: adaptive surface scan (spec detection + heuristics)
- `list`: shows discovered endpoints for current session
- `run`: run a request (uses session base when available)
- `session`: shows active session state

Early project. Feedback welcome.
