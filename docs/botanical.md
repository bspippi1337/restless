# RESTLESS 42 Botanical Harvest (experimental)

A deliberately conservative documentation crawler. Its purpose is to collect API specifications, developer guides, public white papers, and standards **without guessing endpoints or fuzzing**.

## Usage

Preview only; **no network traffic**:

```sh
restless harvest https://docs.example.org/ --host docs.example.org
```

Fetch only when you have checked the site's terms, licensing, and authority to access the material:

```sh
restless harvest https://docs.example.org/ --host docs.example.org \
  --max-pages 5 --max-depth 1 --delay 3s \
  --out .restless/botanical --execute --authorized
```

Produces an append-only `index.jsonl` with timestamp, source URL, source page, status, MIME type, SHA-256, document size, and an archive filename under `sha256/`. Repeated runs may append duplicate index entries; content hashes deduplicate filenames, not log records. PDFs are stored as bytes, **not yet text-extracted**.

## Safe-by-default rules

- Nothing accesses the network without both `--execute` and `--authorized`. The default command is a dry-run.
- Explicit `--host` must match the seed's hostname, and every discovered URL must use the *same origin*, including scheme and port.
- Production requests use HTTPS only, reject credentials and query-bearing URLs, reject local/private/reserved destination IP addresses, do not use ambient proxies, and do not follow redirects.
- Read `/robots.txt` first. Honor its applicable allow/disallow groups and optional crawl-delay. A missing file (404/410) permits crawling under RFC 9309; other statuses or a retrieval failure stop the crawl. The implementation is conservative and does not claim exhaustive RFC 9309 percent-encoding coverage yet.
- Maximum **total page requests** is 25; default 5. Depth is at most 3; default 1. Request pacing is at least 2 seconds. Each response has a 2 MiB archive cap.
- Stop immediately on HTTP 429 or 5xx (no automatic retry). Ignore other failed page responses.
- Follow only document-like HTML links, never arbitrary API operation URLs. No login, browser automation, active probing, hidden-path guessing, or fuzzing.

`robots.txt` is **not a legal permission grant**. Operators must independently ensure applicable permission, terms of service, privacy rules, copyright/license restrictions, and data protection compliance. Do not collect personal or protected information.

## Current implementation limits

This first slice is a scope-locked **document harvester**, not yet a search engine or OpenAPI operation planner. It recognizes HTML links and archives selected text, JSON, YAML and small PDF resources, but does not parse PDF text, provide incremental ETag updates, or verify license metadata. These are the next stages before production use.

Run focused tests:

```sh
go test ./internal/botanical
```

The integration branch remains a draft PR until formatting, tests, security review, and merge conflict review pass.
