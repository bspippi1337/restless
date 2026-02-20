# Restless
**Domain-first API discovery and interaction engine**  
Evidence-driven **CLI** + interactive **TUI** (Beta)

[![Go Report Card](https://goreportcard.com/badge/github.com/bspippi1337/restless)](https://goreportcard.com/report/github.com/bspippi1337/restless)
[![Latest Release](https://img.shields.io/github/v/release/bspippi1337/restless?color=green)](https://github.com/bspippi1337/restless/releases)
[![License](https://img.shields.io/github/license/bspippi1337/restless)](LICENSE)

Restless automatically explores **real** APIs starting from a **single domain**.  
No OpenAPI spec required.

It maps what actually exists by probing safely, verifying responses, and collecting evidence you can replay later.

---

## What Restless does
Restless is built for the moment where you have **a domain** and **questions**, but no docs.

- **Discovers endpoints from a domain**
- **Verifies** endpoints with real requests (not guesses on paper)
- **Light fuzzing** to tease out hidden resources (without going full chaos)
- **Budgets**: time/page limits so it stays controlled
- **Profiles**: saves discovery results and reuse them later
- **Interactive console / TUI** to build, test, replay, and save request snippets
- **Terminal-native**: fast, minimal, and automation-friendly

> No Electron bloat. No YAML hell.  
> Just a sharp tool that lives where you work: the terminal.

---

## Who it’s for
Perfect for:

- Security researchers / bug bounty hunters probing undocumented APIs
- Reverse engineers analyzing mobile or web backends
- Developers quickly exploring third-party services without docs
- Pentesters building request chains and proofs fast

---

## Quickstart

### 1) Install
Grab a prebuilt release from GitHub Releases:
- https://github.com/bspippi1337/restless/releases

Example install (Linux/macOS):
```bash
curl -L https://github.com/bspippi1337/restless/releases/latest/download/restless_Linux_x86_64.tar.gz | tar xz
sudo mv restless /usr/local/bin/
restless --version
