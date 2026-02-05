<p align="center">
  <img src="brand/restless-banner.png" alt="restless banner" width="100%">
</p>

<p align="center">
  <img src="brand/restless-logo.png" alt="restless logo" width="140">
</p>

<h1 align="center">restless</h1>
<p align="center"><b>Universal API Client</b> (wizard-first, profile-driven, secure-by-default)</p>

<p align="center">
  <a href="https://github.com/blckswan1337/restless/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/blckswan1337/restless/ci.yml?branch=master"></a>
  <a href="https://github.com/blckswan1337/restless/releases"><img alt="Release" src="https://img.shields.io/github/v/release/blckswan1337/restless"></a>
</p>

---

## What it is

`restless` is a Go-based **universal API client** designed for fast, repeatable interactions with unfamiliar APIs:

- **Connect & Discover wizard**: start with a domain, end with a usable profile
- **Domain-based API detection**: best-effort guessing of “what am I talking to?”
- **Profiles & presets**: save endpoints, auth strategy, headers, rate limits, timeouts
- **Pluggable auth strategies**: bearer, basic, api-key, OAuth-style flows (extensible)
- **Secure secret handling**: keep tokens out of shell history and plain-text files
- **CLI-first output**: human + JSON, easy to pipe into other tools

> Note: The repository started as a bootstrap skeleton. This update adds packaging + release automation and a consistent visual brand.

## Install

### Option A: Download a release binary
Grab the latest from GitHub Releases:
- https://github.com/blckswan1337/restless/releases

### Option B: npm (wrapper that downloads the matching release binary)
After you publish the npm package (see `PUBLISHING.md`):
### Option C: Homebrew (macOS/Linux)
After you create a Homebrew tap and publish a release:

```bash
brew tap blckswan1337/tap
brew install restless
```


```bash
npm i -g restless-uac
restless --help
```

## Build (local)

```bash
go test ./...
go build ./cmd/uac
```

## Contributing

- Keep changes small and testable
- Prefer adding new auth strategies as isolated packages under `internal/`

## Brand assets

See `brand/BRAND.md` for usage rules.
