# <svg width="28" height="28" viewBox="0 0 24 24" style="vertical-align:middle"><path fill="black" d="M12 2c-2 2-3 4-3 6 0 2 1 3 2 4-2 1-3 3-3 5 0 3 3 5 6 5s6-2 6-5c0-2-1-4-3-5 1-1 2-2 2-4 0-2-1-4-3-6-1 1-2 2-4 2z"/></svg> Restless

**Find the truth of an API. Fast.**

Restless is a terminal-first API reconnaissance tool that **discovers endpoints**, **infers structure**, and **maps topology**.  
It can also **detect OpenAPI / Swagger automatically** and use that signal to improve discovery.

Inspired by tools like `ripgrep`, Restless is built around a simple philosophy:

- Minimal friction
- Immediate output
- Reality over documentation

Point it at an API and start exploring.

---

## Install

### Build from source

```bash
git clone https://github.com/bspippi1337/restless
cd restless
make build
```

Binary:

```bash
./build/restless --help
```

Optional install:

```bash
sudo make install
```

---

## Quickstart

```bash
./build/restless scan https://api.github.com
./build/restless map https://api.github.com
./build/restless inspect GET /users/{user}
```

Restless stores scan state locally:

```
~/.restless_state.json
```

This means you can scan once and iterate without rescanning.

---

## Terminal demo

### Scan

```bash
./build/restless scan https://api.github.com
```

Example output:

```text
Saved scan → ~/.restless_state.json
Routes discovered: 1
OpenAPI detected
```

---

### ASCII API topology

```bash
./build/restless map https://api.github.com
```

Example:

```text
https://api.github.com
├── users
│   ├── /{user}
│   └── /{user}/repos
├── repos
│   ├── /{owner}/{repo}
│   └── /{owner}/{repo}/issues
└── search
```

This output is designed to be:

- readable in terminals
- pasteable into tickets
- usable in architecture notes
- usable in incident reports

---

### Inspect an endpoint

```bash
./build/restless inspect GET /users/{user}
```

Example output:

```text
METHOD
  GET

PATH
  /users/{user}

SOURCE
  discovered during scan
```

---

## OpenAPI / Swagger autodetect

During scanning, Restless probes common schema locations such as:

```
/openapi.json
/swagger.json
/api-docs
/docs/openapi
/v1/swagger
```

If a schema is found, Restless uses it to:

- improve endpoint discovery
- confirm route structure
- detect documentation drift

This allows Restless to operate effectively even when:

- documentation is incomplete
- gateways obscure endpoints
- production APIs differ from spec

---

## Commands

```
restless scan <url>        Discover endpoints and persist scan state
restless map <url>         Render API topology as an ASCII tree
restless inspect <M> <P>   Inspect a specific endpoint
restless discover <url>    Additional discovery mode
restless swarm <url>       Distributed probing
restless magiswarm <url>   Recon engine: discover, fuzz, map, report
restless blckswan <url>    Full recon pipeline
restless octoswan <url>    Parallel probing + inference
```

Use:

```
restless --help
```

for the currently available commands.

---

## Example use cases

### Explore an unknown API surface

```bash
./build/restless scan https://internal-api.company
./build/restless map https://internal-api.company
```

---

### Visualize an API structure quickly

```bash
./build/restless map https://api.github.com
```

---

### Investigate a specific endpoint

```bash
./build/restless inspect GET /repos/{owner}/{repo}
```

---

## Swagger workflow generator

For teams maintaining OpenAPI schemas, a useful extension is automated schema validation in CI.

Future workflow generator concept:

```bash
restless workflow github --openapi-guard https://api.example.com
```

This would generate:

```
.github/workflows/restless-openapi-guard.yml
```

The workflow could:

- fetch the OpenAPI schema
- compare against stored snapshots
- detect undocumented endpoints
- fail CI on breaking drift

This transforms Restless into an **API integrity guardrail**, not just a mapper.

---

## Project layout

```
cmd/restless/        CLI entrypoint
internal/cli/        command definitions
internal/core/       discovery and probing engine
docs/                documentation
examples/            example configurations
assets/              graphics and demo media
archive/             legacy experiments and history
```

---

## Design principles

- Terminal-first interface
- Fast feedback loops
- Minimal dependencies
- Accurate representation of real APIs

---

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
make build
```

---

## License

MIT
