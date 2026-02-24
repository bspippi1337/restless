# Restless v2 Roadmap

## Strategy
Build a stable v2 core with compile-time modules (single binary), then expand features.
Avoid runtime plugins until the core is boring and stable.

## Principles
- Core is sacred: small, stable, testable.
- Modules evolve fast but must respect boundaries.
- CLI first, TUI premium, GUI last.
- Every phase has exit criteria.

---

## Phase 0: Baseline
- gofmt, tidy, go test ./...
- remove stale entrypoints or isolate behind build tags
- docs: architecture + roadmap aligned with code

Exit:
- `go test ./...` is green

## Phase 1: Engine-first core
- Stable request/response types
- Runner interface
- HTTP runner default
- App wiring surface + hooks (request/response mutators)

Exit:
- CLI can run requests using core app

## Phase 2: Sessions v1
- {{vars}} templating in url/body/headers
- simple JSON dot-path extractor
- regex extractor

Exit:
- example flow demonstrates "extract then reuse"

## Phase 3: OpenAPI v1
- import + cache specs
- list cached specs
- (next) parse endpoints + quick-run

Exit:
- deterministic import/list behavior and docs

## Phase 4: Bench v1
- concurrency run + p50/p95/p99
- output table + json export

Exit:
- bench works reliably, warnings for throttling

## Phase 5: Export/Artifacts
- json artifact saving
- md/html report later

Exit:
- one command makes shareable artifact

## Phase 6: TUI 2.0
- tabs, history, request builder, json viewer

Exit:
- feels like a product (lazygit/k9s vibe)

## Phase 7: GUI
- minimal shell using same core

Exit:
- GUI shares core 100%

## Future: Plugin runtime (WASM)
- define stable plugin interfaces first
- loader later
