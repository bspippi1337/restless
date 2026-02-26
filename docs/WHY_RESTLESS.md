# Why Restless

Restless exists because most API tooling optimizes for interaction.
Restless optimizes for execution.

It is not a GUI replacement.
It is not a curl wrapper.
It is not a load-testing suite.

It is a deterministic, OpenAPI-first execution engine.

---

## The Problem With Existing Tools

### curl

- Stateless
- No contract awareness
- No schema validation
- No environment abstraction
- No execution memory

You build everything manually every time.

### Postman / Insomnia

- GUI state-driven
- Difficult to diff or version properly
- Hard to embed in CI
- Encourages manual workflows
- Contract drift can hide silently

### k6 / load testing tools

- Great for load testing
- Not designed for contract-first validation
- Require additional scripting layer

---

## What Restless Does Differently

### 1. OpenAPI As Executable Contract

Restless treats the spec as an execution boundary.

- Path must exist
- Method must match
- Parameters must be present
- Strict mode enforces failure

This eliminates “it worked locally” ambiguity.

---

### 2. Deterministic Execution

Every command:

- Has a predictable exit code
- Has no hidden state
- Can be run in CI
- Produces stable output

There is no GUI memory.
No hidden variables.
No interactive-only state.

---

### 3. Profile-Based Environments

Instead of duplicating requests across dev/stage/prod:

    restless profile set prod base=https://api.example.com
    restless profile use prod

Base injection happens automatically.

This makes environment switching explicit and reproducible.

---

### 4. Scriptability First

Restless works with:

- jq
- fzf
- cron
- CI pipelines
- Makefiles
- shell scripts

Example:

    restless openapi run <id> GET /health | jq '.status'

No export step required.
No conversion step required.

---

### 5. CI-Native Behavior

With strict mode:

    export RESTLESS_STRICT=1

If a path disappears from the spec,
if a method changes,
if a parameter is missing,

Execution fails immediately.

This turns OpenAPI into a contract gate.

---

## When To Use Restless

Use Restless when:

- You want contract-aware execution
- You want deterministic CLI workflows
- You want OpenAPI validation in CI
- You want reproducible environment handling
- You prefer composable shell pipelines

---

## When Not To Use Restless

Do not use Restless when:

- You need visual API exploration for non-technical users
- You want drag-and-drop workflow builders
- You want full GUI collection management

Restless is intentionally not that.

---

## Philosophy

Restless is built around five principles:

1. Determinism over convenience
2. Contracts over assumptions
3. CLI over GUI
4. Explicit configuration over hidden state
5. Reproducibility over interactivity

---

## Positioning

Restless sits between:

curl  → raw execution  
Postman → interactive GUI  
k6 → load testing  

It focuses on contract-aware execution and automation.

---

Restless is not designed to replace everything.

It is designed to make API execution reliable.
