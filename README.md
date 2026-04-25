# RESTLESS

<p align="center">
<b>API topology inference CLI</b><br>
Discover • Model • Inspect • Explain API surfaces from the terminal
</p>

---

## Overview

Restless is a command-line tool for API topology inference. It discovers endpoints, models API structure, detects schema hints, renders graphs, and stores session state so follow-up commands can work without repeating the target.

Restless is designed for developers, maintainers, API reviewers, and security-conscious operators who need transparent, bounded, and reproducible API exploration.

---

## Features

- Bounded same-host API discovery
- Safe probing with GET, HEAD, and OPTIONS
- Deterministic command output
- Session-aware workflows
- XDG-compliant state storage
- JSON schema field inference
- Graph output for topology visualization
- No telemetry or background analytics
- Release-build friendly version metadata

---

## Installation

```bash
curl -sSL https://bspippi1337.github.io/restless/install.sh | sh
```

or:

```bash
wget -qO- https://bspippi1337.github.io/restless/install.sh | sh
```

From source:

```bash
git clone https://github.com/bspippi1337/restless.git
cd restless
go build -o restless ./cmd/restless
```

Release metadata can be injected with Go linker flags:

```bash
go build -o restless ./cmd/restless \
  -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=0.4.0 \
            -X github.com/bspippi1337/restless/internal/version.Commit=$(git rev-parse --short HEAD) \
            -X github.com/bspippi1337/restless/internal/version.Date=$(date -u +%Y-%m-%d)"
```

---

## Quick start

```bash
restless scan https://api.github.com
restless map
restless inspect GET /users
restless graph
restless teach
restless copilot
```

Restless stores the latest scan as session state. In normal interactive use, you only need to provide the target once.

---

## Core workflow

```text
scan → learn → map → inspect → graph → teach → copilot
```

| Command | Purpose |
|--------|---------|
| `scan` | perform fast endpoint discovery |
| `learn` | run documentation-aware discovery |
| `map` | print a structural endpoint overview |
| `inspect` | inspect a method and path from the current session |
| `graph` | render API topology |
| `teach` | explain the inferred API structure |
| `copilot` | suggest useful next commands |
| `fuzz` | probe common API paths heuristically |
| `engine` | run the discovery engine directly |
| `version` | print build and platform metadata |
| `gnu` | print a free-software friendly greeting |

---

## Example session

```bash
restless scan api.example.com
restless map
```

Example output:

```text
/
├── users
├── users/{id}
│   └── posts
└── health
```

Then inspect a route:

```bash
restless inspect GET /users
```

Example output:

```text
Route found
Example:
curl https://api.example.com/users
```

---

## Session state

Restless stores session state using the XDG state directory:

```text
$XDG_STATE_HOME/restless/state.json
```

If `XDG_STATE_HOME` is not set, Restless uses:

```text
~/.local/state/restless/state.json
```

For compatibility, Restless can still read the legacy state file:

```text
~/.restless_state.json
```

---

## Network model

Restless is intentionally conservative during discovery:

- it only follows links on the original target host
- it uses bounded traversal depth
- it limits response reads
- it does not send mutation requests during discovery
- it does not perform telemetry or background network calls

Discovery uses safe HTTP methods only:

```text
GET
HEAD
OPTIONS
```

---

## Graph output

Generate a topology graph:

```bash
restless graph https://api.example.com
```

Generate DOT output:

```bash
restless graph https://api.example.com --format dot
```

---

## Packaging notes

Restless is intended to be distribution-friendly:

- no generated binary blobs are required in the source tree
- version metadata can be injected at build time
- command output is deterministic where practical
- session state follows the XDG base directory model
- offline fixtures can be used for test-oriented workflows

---

## Project structure

```text
cmd/restless        main command entrypoint
internal/cli        command definitions
internal/core       core scan, state, and utility packages
internal/discovery  topology discovery engine
internal/version    release metadata
man/                manual page sources
testdata/           offline fixtures
```

---

## Design principles

- Discovery first
- User-controlled operation
- Transparent network behaviour
- Graph-based API understanding
- Reproducible command output
- Free-software friendly defaults

---

## License

MIT
