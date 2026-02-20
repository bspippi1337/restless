# Restless

**Domain-first API discovery and interaction engine**  
Evidence-driven CLI + interactive TUI (Beta)

[![Go Report Card](https://goreportcard.com/badge/github.com/bspippi1337/restless)](https://goreportcard.com/report/github.com/bspippi1337/restless)
[![Latest Release](https://img.shields.io/github/v/release/bspippi1337/restless?color=green)](https://github.com/bspippi1337/restless/releases)
[![License](https://img.shields.io/github/license/bspippi1337/restless)](LICENSE)

Restless automatically explores real APIs starting from a single domain — no OpenAPI spec required.  
It verifies endpoints, fuzzes lightly for hidden resources, respects budgets (time/pages), saves profiles, and lets you jump into an interactive console to build, test, replay, and save request snippets.

Perfect for:

- Security researchers / bug bounty hunters probing undocumented APIs
- Reverse engineers analyzing mobile/web backends
- Developers quickly exploring third-party services without docs
- Pentesters building request chains fast

No Electron bloat. No YAML hell. Just fast, terminal-native discovery and interaction.

## Quickstart

Install the latest release (or build from source — see below):

```bash
# Example: Linux/macOS (amd64/arm64)
curl -L https://github.com/bspippi1337/restless/releases/latest/download/restless_Linux_x86_64.tar.gz | tar xz
sudo mv restless /usr/local/bin/
