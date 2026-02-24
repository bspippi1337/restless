#!/usr/bin/env bash
set -euo pipefail

VERSION="4.0.4"

echo "==> Upgrading documentation with real examples and deep dives"

# ----------------------------------
# README Upgrade
# ----------------------------------

cat > README.md <<EOT
# Restless

Terminal-First API Workbench  
Version: $VERSION

Restless is a modular OpenAPI-aware execution engine designed for deterministic, scriptable API interaction.

---

## Quick Example (Real Output)

Import spec:

    restless openapi import petstore.json

Output:

    imported: 3cd1f17961f8e4ebe06369a6fd15699195ecda0f

Run endpoint:

    restless openapi run 3cd1f17961f8e4ebe06369a6fd15699195ecda0f GET /pets

Output:

    status: 200 (dur=124ms)
    ---
    [
      {
        "id": 1,
        "name": "doggie",
        "status": "available"
      }
    ]

---

## Strict CI Mode

    export RESTLESS_STRICT=1
    restless openapi run <id> GET /missing

Output:

    ERROR: path not found in spec

---

## Philosophy

Restless is:

- Deterministic
- CI-native
- Scriptable
- Modular
- OpenAPI-first

See /docs for deep technical reference.
EOT

# ----------------------------------
# WORKFLOWS
# ----------------------------------

cat > docs/WORKFLOWS.md <<'EOT'
# Workflows & Integrations

## With jq

Pipe JSON directly:

restless openapi run <id> GET /pets | jq '.[] | .name'

## With fzf

Select endpoint interactively:

restless openapi endpoints <id> | fzf

## With curl generation

restless openapi run <id> GET /pets --curl

Use in scripts or debugging.

## With CI pipelines

export RESTLESS_STRICT=1
restless openapi run <id> GET /health

Exit code will fail pipeline if spec mismatch.

## With cron / automation

restless openapi run <id> GET /metrics --save metrics_snapshot

Artifacts stored deterministically.
EOT

# ----------------------------------
# OPENAPI DEEP DIVE
# ----------------------------------

cat > docs/OPENAPI_DEEP_DIVE.md <<'EOT'
# OpenAPI Deep Dive

Restless does not treat OpenAPI as documentation.
It treats it as an executable contract.

## Import

restless openapi import spec.yaml

Spec is parsed and indexed locally.

## Endpoint Resolution

- Path + method validated
- operationId supported
- Path parameters enforced
- Interactive prompting when missing

Example:

restless openapi run <id> GET /pets/{petId}

If petId missing:

Enter petId: 7

## Profile Base Injection

If no --base specified:

Active profile base is injected automatically.

restless profile set prod base=https://api.example.com
restless profile use prod

Now:

restless openapi run <id> GET /pets

Uses https://api.example.com automatically.

## Validation Layer

- Path existence
- Method validity
- Missing path params
- Strict CI enforcement

This makes Restless spec-aware and deterministic.

## Curl Generation

restless openapi run <id> GET /pets --curl

Outputs fully constructed curl command.

Useful for debugging or sharing requests.

## Artifact Storage

restless openapi run <id> GET /pets --save snapshot

Stores response JSON locally.

Artifacts are stable and scriptable.
EOT

# ----------------------------------
# CI GUIDE
# ----------------------------------

cat > docs/CI.md <<'EOT'
# CI Integration

Restless is designed to fail deterministically.

## Example GitHub Actions step

- name: API Health Check
  run: |
    export RESTLESS_STRICT=1
    restless openapi import spec.json
    restless openapi run <id> GET /health

If endpoint changes or spec mismatches,
the job fails immediately.

## Why Strict Mode Matters

Without strict mode:
fallback behavior may hide contract drift.

With strict mode:
any contract mismatch stops execution.

This ensures API contract integrity in CI pipelines.
EOT

echo "==> Documentation upgraded successfully"
