# Restless v2 Migration Guide

## What this ZIP adds

- Core engine abstraction
- HTTP adapter
- App runner layer
- CLI wired to engine
- Clean layered architecture

## How to apply

1. Extract ZIP at repo root.
2. Run: go mod tidy
3. Build: go build ./cmd/restless
4. Test: ./restless run GET https://api.github.com

## Next Steps

- Move old command logic into app layer.
- Add history + snapshot modules.
- Introduce plugin interface.
- Remove duplicated legacy paths.
