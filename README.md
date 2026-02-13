[![Release](https://img.shields.io/github/v/release/bspippi1337/restless?include_prereleases&label=latest)](https://github.com/bspippi1337/restless/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bspippi1337/restless)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**restless** is a CLI-first API client that discovers and learns an API from just one input: **a domain name**.

Tired of manually reading through pages of API docs just to find a working endpoint? Type `bankid.no`, press discover, and restless automates the boring parts. It's a disciplined, safety-focused explorer for modern web APIs.

![A conceptual animation of the restless "Path" logo discovering endpoints](assets/brand/logo/restless_logo_B_animated.svg)

## ✨ Key Features

*   **🔭 Domain-First Discovery**: Start with only a domain (e.g., `api.example.com`). Restless handles the rest.
*   **📄 Intelligent Documentation Finding**: Automatically locates OpenAPI specs, developer portals, and common documentation paths.
*   **🧪 Safe & Disciplined Fuzzing**: Generates tests **only from seed words** found in docs or common API patterns. No brute-force, no auth bypass, no dangerous write calls by default.
*   **✅ Endpoint Verification**: Confirms valid endpoints using safe methods (GET, HEAD, OPTIONS) first.
*   **🖥️ Dual Interface**: A powerful terminal UI (TUI) for interactive exploration and a classic CLI for scripting and automation.
*   **🛡️ Safety by Default**: Built with hard-coded request budgets and a "verify, don't attack" philosophy.

## 📦 Installation

### 🚀 Quick Install (Recommended)
Get the latest pre-built binaries for Windows, Linux, and macOS from the [**Actions**](https://github.com/bspippi1337/restless/actions) page or the [**Releases**](https://github.com/bspippi1337/restless/releases) section.

1.  Go to the latest release or successful workflow run.
2.  Download the artifact for your operating system.
3.  Extract and run the `restless` binary.

### 🛠️ Build from Source
If you have Go installed, you can build the latest version directly:
```bash
git clone https://github.com/bspippi1337/restless.git
cd restless
make build
./bin/restless
