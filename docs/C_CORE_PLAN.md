# C Conversion Plan (no vendored deps)

Restless is currently a Go CLI/TUI built around **domain-first discovery** and profiles/snippets. citeturn2view0

A full rewrite to C with **zero external libraries** is mainly blocked by **HTTPS/TLS**:
- Writing TLS from scratch is unsafe.
- The practical “no external deps” approach is: **use OS TLS backends** (WinHTTP/SChannel, SecureTransport/Network.framework, system OpenSSL on Linux/Termux).

## Recommended strategy (fast + safe)

1) **Introduce a C core binary** (`corec/bin/restless-core`) for the engine.
2) Keep the Go CLI/TUI as a frontend (or migrate later). The frontend calls the core via:
   - JSON stdin/stdout protocol, or
   - exec + JSON output (v0).
3) Gradually move modules:
   - profiles/snippets/history to C (simple line-based format)
   - request runner (HTTP/HTTPS) with backend abstraction
   - discovery pipeline (OpenAPI seeds, sitemap, doc pages)
   - fuzz “seed-only” as already described in README citeturn1view0

## “Obscure the source” without sketchiness
- Ship a stripped binary: `-s -Wl,--strip-all`, LTO, hidden symbols.
- Keep the core in C and the UI in Go if you want the repo to stay readable while the “engine” is a black box.

## Next step
Implement HTTPS backend per platform and wire Go `discover` to call `corec/bin/restless-core discover <domain> --json`.
