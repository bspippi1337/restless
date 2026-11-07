#!/usr/bin/env bash
set -euo pipefail

VERSION="4.0.4"

echo "==> Rewriting README for v$VERSION"

cat > README.md <<EOT
# Restless ⚡

**Terminal-First API Workbench**

Restless is a modular, OpenAPI-aware execution engine built for developers who live in the shell.

Version: **$VERSION**

---

## ✨ Highlights

- OpenAPI import (JSON + YAML)
- Endpoint discovery
- OperationId execution
- Interactive path parameter prompt
- Environment profiles
- Session variable templating
- Curl generation
- Artifact export
- Built-in benchmarking
- Latency histogram
- Strict mode for CI
- Cross-platform static builds

---

## 🚀 Install

### From source

\`\`\`bash
go build -o restless ./cmd/restless
\`\`\`

### Download binary

See GitHub Releases → v$VERSION

---

## 📦 OpenAPI Workflow

### Import spec

\`\`\`bash
restless openapi import petstore.json
\`\`\`

### List specs

\`\`\`bash
restless openapi ls
\`\`\`

### List endpoints

\`\`\`bash
restless openapi endpoints <id>
\`\`\`

---

## ▶ Run endpoint

\`\`\`bash
restless openapi run <id> GET /pets
\`\`\`

Path parameters auto-prompt if missing.

With explicit param:

\`\`\`bash
restless openapi run <id> GET /pets/{petId} -p petId=7
\`\`\`

Generate curl:

\`\`\`bash
restless openapi run <id> GET /pets --curl
\`\`\`

---

## 🌍 Profiles

Set base URL:

\`\`\`bash
restless profile set dev base=https://petstore3.swagger.io/api/v3
restless profile use dev
\`\`\`

List profiles:

\`\`\`bash
restless profile ls
\`\`\`

---

## 🔐 Session variables

\`\`\`bash
restless openapi run <id> GET /secure \
  -H "Authorization: Bearer {{token}}" \
  -set token=abc123
\`\`\`

---

## 📈 Benchmark

\`\`\`bash
restless -url https://httpbin.org/get -bench
\`\`\`

Includes latency percentiles and histogram.

---

## 🛡 Strict Mode

Fail hard for CI:

\`\`\`bash
export RESTLESS_STRICT=1
\`\`\`

---

## 🏗 Architecture

- core/app → module registry
- modules/openapi → spec engine
- modules/session → templating
- modules/export → artifacts
- modules/bench → performance
- internal/version → centralized version injection

---

## 🎯 Philosophy

Restless is designed to be:

- Deterministic
- Scriptable
- CI-friendly
- Terminal-native
- Modular

Not a GUI wrapper.
A composable execution engine.

---

## 🏷 Release

This repository currently tracks version **$VERSION**.

See GitHub Releases for binaries and checksums.

---

## 📜 License

MIT
EOT

echo "==> README updated"
