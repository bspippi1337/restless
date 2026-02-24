# Restless v2 Roadmap

## Mål
Bygge Restless v2 som en stabil, terminal-first API workbench:
- Én binær
- Modulær kodebase (compile-time modules)
- OpenAPI + sessions + TUI i toppklasse
- Klar for fremtidig plugin-system (WASM) uten å forplikte oss nå

---

## Guiding Principles
- Core skal være liten, stabil og testbar.
- Moduler kan itereres raskt, men må ha klare grenser.
- CLI UX først. TUI er premium. GUI kommer når core er “boring”.
- Shipping > perfection, men hver release må være solid.

---

## Phase 0: Repo hygiene + Baseline (1. uke)
**Deliverables**
- Rydding: døde filer, “bak”, støy, gamle entrypoints
- Standardisering: gofmt, golangci-lint, test workflow
- Single source of truth: `cmd/restless` er eneste entry
- Stabil README (kort, korrekt, ingen døde linker)

**Exit criteria**
- `go test ./...` grønt
- `go vet ./...` grønt
- CI bygger release artifacts
- README beskriver faktisk hva programmet gjør nå

---

## Phase 1: Core refactor (Engine-first) (1–2 uker)
**Deliverables**
- Ny struktur:
  - `internal/core/{config,httpx,engine,store,types}`
  - `internal/modules/*` (tomt i starten er ok)
- Request execution pipeline:
  - request model
  - response model
  - middleware hooks (auth, headers, logging)
- History store (min.):
  - siste N requests/responses
  - lagres i `~/.restless/` (json/sqlite senere)

**Exit criteria**
- CLI kan kjøre requests stabilt
- Enhetstester på engine (happy path + fail path)
- Ingen modul importerer core feil vei (core -> modules forbudt)

---

## Phase 2: Session Engine v1 (1–2 uker)
**Deliverables**
- Variabler:
  - `{{var}}` template i headers/body/url
- Extractors v1:
  - JSONPath (eller enkel dot-notation først)
  - Regexp extractor
- “Chain”:
  - kjør request A, extract, kjør request B

**CLI UX**
- `restless run flow.yaml`
- `restless set token=...`
- `restless env use prod|stage`

**Exit criteria**
- Demo-flow i `examples/` som kjører end-to-end
- Dokumentasjon: “Sessions for dummies”
- Stabil feilhåndtering (gode error messages)

---

## Phase 3: OpenAPI Mode v1 (2–3 uker)
**Deliverables**
- Import OpenAPI 3.x (URL eller fil)
- Cache schema lokalt
- Endpoint listing + quick-run:
  - velg endpoint
  - autofyll base url
  - param prompts (minimal)
- Snippet export:
  - curl
  - httpie (bonus)
  - go (bonus)

**Exit criteria**
- `restless openapi import <url>`
- `restless openapi ls`
- `restless openapi run /path --method GET`
- README viser ekte OpenAPI workflow

---

## Phase 4: TUI 2.0 (2–4 uker)
**Deliverables**
- Bubble Tea “pro feel”:
  - request builder view
  - response viewer (JSON tree + raw)
  - history timeline
  - tabs (Requests/History/OpenAPI/Flows)
- Keymap + help overlay
- Theme baseline (min 2)

**Exit criteria**
- TUI føles som et produkt (lazygit/k9s-nivå i flyt)
- Ingen crashes ved store responses
- Responsiv på langsomme nett (loading states)

---

## Phase 5: Bench Mode v1 (1–2 uker)
**Deliverables**
- Concurrency test:
  - `restless bench <url> -c 20 -d 10s`
- Metrics:
  - p50/p95/p99
  - error rate
  - throughput
- Output:
  - terminal table
  - json export

**Exit criteria**
- Bench er deterministisk nok til CI-run
- God warning ved rate limits / throttling

---

## Phase 6: Export + Artifact System (1–2 uker)
**Deliverables**
- Export response:
  - json
  - md report
  - (bonus) html report
- “Artifacts” mappestruktur:
  - `~/.restless/artifacts/<date>/<name>/`

**Exit criteria**
- Én kommando som genererer en fin rapport du faktisk kan sende videre

---

## Phase 7: GUI (Når core er stabil)
**Deliverables**
- Minimal GUI shell som bruker samme engine:
  - saved environments
  - flow runner
  - request builder
- Fokus: stabilitet og integrasjon, ikke 200 features

**Exit criteria**
- GUI deler 100% core med CLI/TUI
- Ingen GUI-only logikk i core

---

## Phase 8: Plugin-ready design (Future)
**Goal**
- Design plugin API (WASM-vennlig) uten å implementere loader ennå.

**Deliverables**
- Stable interfaces:
  - request middleware hooks
  - exporters
  - auth providers
- Document: `docs/plugin-api.md`

**Exit criteria**
- Core-grenser er tydelige
- Ingen “lekkasje” av internstruktur i API

---

## Release plan
- v2.0.0: Core + Sessions + OpenAPI v1 + TUI 2.0 baseline
- v2.1.0: Bench + Export + artifacts
- v2.2.0: GUI shell
- v2.3.0+: Plugin runtime (WASM) vurderes

---
