# RESTLESS 42 Botanical (experimental)

A conservative offline-first knowledge and discovery pipeline for documented API surfaces:

**Harvest → verify → index → search → plan**.

The goal is to answer "what does this API already support?" from official documentation instead of trying unlisted paths or fuzzing. This is an experimental component on the draft integration branch.

## Quick start

```sh
go build -o bin/restless ./cmd/restless
DIR="$HOME/.local/share/restless-42/botanical"

# Network-free preview (default)
bin/restless harvest https://petstore3.swagger.io/api/v3/openapi.json --host petstore3.swagger.io --out "$DIR"

# One explicitly approved, public demo specification
# Only after independently checking access rights, terms, licensing and purpose.
bin/restless harvest https://petstore3.swagger.io/api/v3/openapi.json \
  --host petstore3.swagger.io --out "$DIR" \
  --max-pages 1 --max-depth 0 --execute --authorized

# These three operations are OFFLINE: no network requests.
bin/restless index --dir "$DIR"
bin/restless search "find pets by status" --dir "$DIR" --limit 5
bin/restless refs "petstore" --dir "$DIR" --limit 5
bin/restless plan "find pet by id" --dir "$DIR" --limit 3

# Compare previously archived versions without contacting the server
# Keep two separate archive directories and run "index" in each first.
bin/restless diff --before "$HOME/restless-archives/v1" --after "$HOME/restless-archives/v2"
```

## Archive and provenance

- Harvest saves documents under `sha256/<digest>` and appends URL, content type, depth, source and SHA-256 provenance to `index.jsonl`.
- Index rereads local archives, verifies SHA-256 for **every document**, and rejects invalid provenance/path traversal before indexing anything.
- `catalog.json` is atomically replaced and stores documents alongside explicitly documented OpenAPI 3 HTTP operations. It includes method, path, summary, tags, required parameters, declared security requirements, source URL and hash.
- Duplicate URLs in the manifest take the last entry. No network access occurs during indexing.
- Plain text / HTML / Markdown are indexed as searchable reference terms, exposed through `restless refs`. PDF bytes are marked `pdf-unparsed`; PDF text extraction is **not implemented**. Neither arbitrary prose nor PDF creates API endpoints.
- Only OpenAPI **3.x** paths are recognized currently. Local parameter references under `#/components/parameters` are resolved offline with cycle and depth checks; external, missing or cyclic parameter references are flagged for manual review. Referenced path items are skipped, and unresolved request-body references are flagged; schema resolution and server selection are not implemented.

## Offline reference search, operation search and probe planning

`restless refs` searches local documentation records and returns source URL, document type, content hash and matching terms. `restless search` uses deterministic weighted lexical ranking of documented operations (operation ID, method, path, summary, tags, description, parameter names). It is not embedding-based or LLM semantic search yet.

`restless plan` returns ranked candidates and evidence links. `restless diff` compares operation contracts across two locally indexed snapshots and reports additions, removals, and modifications; it ignores changes to archive hashes when the documented operation itself is unchanged. Each candidate includes its documented method/path, source hash, declared authentication, required parameters and review warnings. Non-read methods, credentials, path parameters and unresolved references demand manual review. **Plan never makes requests, runs scripts or resolves URLs.** The JSON field `network_requests` is always zero.

A plan is not approval to call an API. A documented GET may have real side effects; explicit operator authorization and target scoping remain mandatory for later live probing.

## Harvester safety limits

- Dry-run by default. Both `--execute` and `--authorized` are required to enable network access.
- Explicit `--host` must match seed hostname; discovered URLs use the same scheme, hostname and port.
- Production enforces HTTPS, no credentials or query-bearing URLs, rejects private/local/reserved IP destinations, avoids ambient proxies, and never follows redirects.
- Read robots.txt first. A 404 or 410 permits crawling under RFC 9309, otherwise unavailable/blocked robots stops collection. Optional `crawl-delay` is honored. The robots parser does not claim exhaustive RFC 9309 percent-encoding conformance.
- Maximum 25 total page attempts; default 5. Depth at most 3; default 1. At least 2 seconds between page requests. Maximum response archive size 2 MiB.
- HTTP 429 and 5xx stop the crawl without retries. Only documentation-like links are followed, no hidden endpoint guessing, login, browser automation, or fuzzing.
- robots.txt is not permission: honor site policy, applicable law, copyright, licensing, privacy and data protection.

## Tests

```sh
go test ./internal/botanical ./internal/cli
go vet ./...
```

This draft PR must not be merged to main before tests, documentation review, and scope/security review pass.
