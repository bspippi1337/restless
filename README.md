# RESTLESS 42

**Documentation-first API discovery and a cross-platform developer workbench.**

[![Go CI](https://github.com/bspippi1337/restless/actions/workflows/ci.yml/badge.svg)](https://github.com/bspippi1337/restless/actions/workflows/ci.yml)
[![Pages](https://github.com/bspippi1337/restless/actions/workflows/pages.yml/badge.svg)](https://github.com/bspippi1337/restless/actions/workflows/pages.yml)
[![Installer](https://github.com/bspippi1337/restless/actions/workflows/installer.yml/badge.svg)](https://github.com/bspippi1337/restless/actions/workflows/installer.yml)

RESTLESS helps a developer find the API operation they forgot **without guessing at undocumented endpoints**. The Botanical workflow collects permitted documentation, verifies its origin, extracts explicit OpenAPI contracts, and creates offline review plans before any live API testing.

**RESTLESS 42 is the initiative name, not the binary's semantic version.** The Go CLI may still report version `0.4.0` in `internal/version/version.go`.

## Documentation and browser toolbox

- **[Documentation center](https://bspippi1337.github.io/restless/docs/)** (when repository GitHub Pages is enabled)
- **[32-operation browser toolbox](https://bspippi1337.github.io/restless/toolbox/)** (GitHub Pages)
- **[Complete English field guide](docs/RESTLESS_42_Field_Guide_EN.md)** (also available as a rendered web handbook)
- **[Browser toolbox operation manual](docs/RESTLESS_42_Browser_Toolbox_EN.md)**
- **[NotebookLM podcast source and brief](docs/RESTLESS_42_NotebookLM_Podcast_Brief_EN.txt)**
- **[Botanical architecture and safety notes](docs/botanical.md)**
- **[Repository Garden and native maintenance tools](docs/repo-garden.md)**

The browser toolbox is a static, browser-only workbench inspired by CyberChef's recipe model. It offers JSON, CSV, encoding, SHA-256/512, JWT inspection, text processing, URL utilities, and offline OpenAPI analysis. It has no network execution path or backend processing, and its operation chaining and import/export run locally.

## Build from source

Requirements: **Go 1.24+**, Git, and a supported Go platform.

```sh
git clone https://github.com/bspippi1337/restless.git
cd restless
go build -o build/restless ./cmd/restless
./build/restless --help
```

For Debian/WSL, install Git and Go via your package manager or the official Go distribution before building.

## First safe workflow

The first command is a dry run and makes **no network requests**:

```sh
./build/restless harvest \
  https://petstore3.swagger.io/api/v3/openapi.json \
  --host petstore3.swagger.io \
  --out .restless/botanical
```

Only after confirming authorization, scope, site terms, and licensing may you explicitly enable an actual documentation fetch:

```sh
./build/restless harvest \
  https://petstore3.swagger.io/api/v3/openapi.json \
  --host petstore3.swagger.io \
  --out .restless/botanical --max-pages 1 --max-depth 0 \
  --execute --authorized
```

Then inspect locally without sending additional API requests:

```sh
./build/restless index --dir .restless/botanical
./build/restless refs "petstore" --dir .restless/botanical
./build/restless search "find pet by id" --dir .restless/botanical
./build/restless plan "find pet by id" --dir .restless/botanical
```

For two separate saved documentation snapshots, use `restless diff --before DIR --after DIR` to compare documented operations. Source URLs and hashes remain part of the evidence trail.

## Safety and truth in documentation

Botanical's remote retrieval is host-scoped, robots-aware, rate-limited, size-limited, and opt-in. It records SHA-256 digests and supports conditional revalidation via ETag and Last-Modified. The offline planner executes **zero HTTP requests**; it does not verify deployment status or authenticate an API.

Browser OpenAPI analysis supports **JSON-format OpenAPI 3.x**, not YAML, full schema resolution, or live probes. JWT inspection **does not verify the token's signature**. These are explicit product boundaries, not hidden limitations.

Other legacy RESTLESS commands, including `call`, `watch`, and `fuzz`, are separate features and may have different behavior and effects. Inspect each command's help and target scope. Only test systems you own or are permitted to access.

## Project structure

| Directory | Purpose |
| --- | --- |
| `cmd/restless`, `internal/` | Go CLI and core runtime |
| `internal/botanical/` | Documentation crawler, archive, index, ranking and planning |
| `site/` | Public documentation portal and local-first browser workbench |
| `tools/pages/` | Documentation renderer templates and site checks |
| `tests/toolbox.test.mjs` | Browser transformation regression tests using Node |
| `cmd/restless-doctor/` | Shell-free Go health and repository audit tool |
| `tools/c/restless-clean.c` | Safe, dry-run-by-default C11 cleanup utility |
| `docs/` | Technical documentation and user manuals |
| `archive/historical/` | Preserved one-off scripts and obsolete material |

## Verify and maintain

```sh
go test ./...
go vet ./...
node --test tests/toolbox.test.mjs
make garden
make garden-check
```

`make garden` previews cleanup and audits the root without deleting files. `make garden-apply` is an explicitly destructive operation limited to generated output directories.

## License and contribution

RESTLESS is licensed under the [MIT License](LICENSE). See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidance, [SECURITY.md](SECURITY.md) for the security policy, and [GitHub Issues](https://github.com/bspippi1337/restless/issues) for bug reports.

**Principle:** documentation first, evidence attached, network requests only with authorization.
