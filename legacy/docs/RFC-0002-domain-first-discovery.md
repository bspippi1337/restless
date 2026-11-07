# RFC-0002: Domain-First Discovery, Fuzzer v2, and Doctor Service
Status: Accepted (v2)
Owner: bspippi1337
Updated: 2026-02-08

## 1. Purpose
Restless v2 must onboard an unknown REST API from **one input only**: a domain.

Example input:
- `bankid.no`

Restless will:
1) derive host candidates,
2) parse public documentation (OpenAPI + doc pages),
3) scrape low-noise hints,
4) fuzz **only from seeds** (never blind brute force),
5) safely probe endpoints (GET/HEAD/OPTIONS) to verify existence,
6) output suggestions that seed requests and presets.

## 2. Non-goals
- No brute force, credential stuffing, bypassing auth, or exploit fuzzing.
- No write operations by default.
- No deep crawling by default without explicit opt-in.

## 3. Core Contract (v2)
### 3.1 One input
The UI/CLI must work with only a domain. Base URL is derived automatically.

### 3.2 Safe by default
- Default methods: GET/HEAD/OPTIONS only
- Hard budgets: max time + max pages + max endpoints per phase
- 401/403 are treated as **existence signals** ("auth required")

### 3.3 Evidence-driven discovery
Every endpoint candidate must carry evidence (source, URL, timestamp, confidence-ish score).

## 4. Pipeline (locked)
Domain → Host candidates → Doc sources → Seed endpoints → Fuzzer expansion → Safe probe verify → Findings

## 5. Host Candidate Strategy (v2)
Given `example.com`, Restless tries:
- `https://example.com`
- `https://api.example.com`
- `https://developer.example.com`
- `https://docs.example.com`
- `https://sandbox.example.com`
- `https://staging.example.com` (lower priority)

(plus light heuristics based on discovered redirects)

## 6. Documentation Sources (v2)
- OpenAPI JSON/YAML candidates:
  - `/openapi.json`, `/swagger.json`, `/api-docs`, `/.well-known/openapi.(json|yaml|yml)`
- `sitemap.xml` for docs pages
- lightweight HTML scraping of likely doc pages

## 7. Fuzzer v2 (seed-only)
The fuzzer expands only from:
- OpenAPI paths
- paths found in docs/html/code blocks
- sitemap-discovered docs pages

Heuristics may add common companions to known paths (e.g. list/detail, health/status/version),
but it must never generate massive brute lists.

## 8. Doctor Service (v2)
Command: `restless doctor`
- cleans old builds/logs/artifacts
- checks repo state (optional)
- validates presets
- prints a report and recommended next commands
Doctor never deletes user secrets or presets without explicit confirmation.

## 9. UX / CLI
- `restless` launches the TUI
- `restless discover <domain>` prints a JSON report (optional `--json`)
- `restless doctor` cleans and reports
- `restless help` shows the full guide; `--help` stays short.

## 10. v3+ Direction (not in v2)
- deeper crawling (opt-in)
- JS bundle hint scanning (opt-in, capped)
- real GUI wrapper (native terminal surface)
- preset encryption/keyring integration
