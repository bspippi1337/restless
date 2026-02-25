# RESTLESS

```
████████
    ████
  ██████
    ████
  ██████
████    
```

**Signal correction for APIs.**  
Terminal-first. Structured. Precise.

---

## What is Restless?

Restless is a terminal-native API workbench built for people who:

- Explore APIs daily
- Debug distributed systems
- Work with OpenAPI specs
- Need repeatable, scriptable API workflows
- Refuse to live inside GUI tools

It is not a curl wrapper.  
It is not a GUI exported to CLI.  

It is a structured engine for:

- Discovery
- Execution
- Correction
- Export

---

## Why Restless Exists

Most API tooling falls into two categories:

1. GUI-heavy and hard to automate
2. Raw CLI tools that lack structure

Restless corrects the signal.

It gives you:

- Structured sessions
- Repeatable workflows
- OpenAPI integration
- Clean exports
- Terminal-native speed

---

# Installation

```bash
go install github.com/bspippi1337/restless/cmd/restless@latest
```

Verify:

```bash
restless --version
```

---

# Core Concepts

Restless works in four phases:

1. Probe – discover surface
2. Import – load OpenAPI or define structure
3. Execute – run requests with state
4. Export – generate artifacts

---

# Practical Usage Examples

## 1. Quick API Exploration

```bash
restless probe https://api.github.com
```

Outputs:

- Available routes
- Status responses
- Basic surface mapping

Use this when encountering an unknown API.

---

## 2. Import OpenAPI Specification

```bash
restless openapi import ./petstore.yaml
```

Now endpoints are structured and accessible.

Run:

```bash
restless list
```

To see discovered endpoints.

---

## 3. Execute a Request

```bash
restless run GET /users
```

With parameters:

```bash
restless run POST /users \
  --body '{"name":"Anders","role":"admin"}'
```

Headers:

```bash
restless run GET /private \
  --header "Authorization: Bearer $TOKEN"
```

---

## 4. Save and Reuse Sessions

```bash
restless session save prod-api
restless session load prod-api
```

This lets you:

- Switch environments
- Store authentication
- Re-run flows safely

---

## 5. Export Results

Generate Markdown report:

```bash
restless export --format md --out report/
```

Generate JSON artifact:

```bash
restless export --format json --out artifacts/
```

Use this for:

- CI pipelines
- Documentation
- Incident reports
- Audit logs

---

# Example: Real Workflow

Scenario: Debugging a failing endpoint in production.

```bash
restless session load prod
restless run GET /orders/17291
restless run GET /orders/17291 --header "X-Debug: true"
restless export --format md --out incident-report/
```

You now have:

- Structured output
- Request history
- Exportable documentation

Without leaving the terminal.

---

# Philosophy

Restless is not decorative.

It is:

- Direct
- Deterministic
- Structured
- Built for operators

The logo reflects this:

A stable structure under corrective force.

---

# Roadmap

- Interactive TUI autocomplete
- Flow runner (multi-step sequences)
- Assertion engine
- Plugin system
- Advanced exporters

---

# Contributing

```bash
git clone https://github.com/bspippi1337/restless
cd restless
go build ./cmd/restless
```

Run tests:

```bash
go test ./...
```

---

# License

MIT
