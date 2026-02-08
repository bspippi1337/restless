# restless (v2 alpha)

A **CLI-first** API client that can learn an API from **one input**: a domain.

Type `bankid.no`, press discover, and Restless does the boring part:
- finds docs (OpenAPI, developer portals, doc pages)
- extracts endpoints
- fuzzes *only from seeds* (safe, disciplined)
- verifies endpoints (GET/HEAD/OPTIONS)
- seeds your next request

## Install
### Build from source
```bash
go mod tidy
make build
./bin/restless
```

### Get a prebuilt app (recommended)
This repo includes GitHub Actions to build release binaries for Windows/Linux/macOS.
- Go to **Actions** → latest run → download artifacts
- Or tag a release (e.g. `v0.2.0-alpha`) and download assets

## Run
```bash
restless
```
- In the wizard: enter a domain (e.g. `openai.com`)
- Press **Ctrl+D** to discover endpoints
- Press `?` for the built-in help tab

## Commands
```bash
restless discover bankid.no --json
restless doctor
restless help
```

## Safety by default
- No brute force
- No auth bypass
- No write calls by default
- Hard budgets

See: `docs/SECURITY.md` and `docs/RFC-0002-domain-first-discovery.md`

## License
MIT
