# About Restless

**Reactive API discovery and Unix automation runtime**

Restless helps developers and operators quickly understand unfamiliar systems through safe, automated exploration. It maps system structure without extensive manual investigation, making it invaluable for DevOps, API testing, and system automation workflows.

## What Restless Does

Restless provides three core capabilities:

- **🔍 Safe API Discovery** – Automatically explore remote APIs using only safe HTTP methods (GET, HEAD, OPTIONS). No mutations, no side effects.
- **🗺️ Topology Mapping** – Discover and visualize the structure of repositories and API services before diving into manual exploration.
- **⚡ Filesystem Automation** – Watch directories for changes and trigger custom commands, enabling reactive workflows and automated testing pipelines.

## Why Restless

Modern development involves working with unfamiliar systems constantly. Rather than manual exploration or heavy orchestration frameworks, Restless offers:

- **Local-first execution** – Everything runs on your machine; no external dependencies
- **Shell interoperability** – Composable command-line tools that work with standard Unix pipelines
- **Minimal assumptions** – Designed for headless systems, containers, and minimal runtime environments
- **Reproducible builds** – Debian-friendly, dependency-light packaging suitable for CI/CD pipelines
- **Inspectable behavior** – Clear output designed for terminals and human understanding

## Design Philosophy

Restless embraces traditional Unix design principles:

> "Do one thing and do it well"

Instead of monolithic orchestration layers, Restless provides focused, composable tools that integrate naturally with existing workflows.

## Use Cases

- **API Exploration** – Understand an unfamiliar REST API before building against it
- **System Auditing** – Detect structural drift and changes in deployed systems
- **Development Workflows** – Watch test directories and automatically re-run tests on file changes
- **CI/CD Integration** – Automate discovery and validation tasks in build pipelines
- **Documentation** – Generate topology maps for system architecture documentation

## Technical Highlights

- **Language:** Go
- **License:** MIT
- **Packaging:** Debian-friendly, reproducibly buildable
- **Distribution:** Suitable for containerized and headless environments
- **Dependencies:** Intentionally minimal

## Getting Started

### Quick Start

```bash
# Explore the current directory structure
restless map .

# Safely scan a remote API
restless scan https://example-api.dev

# Watch a directory and run tests on changes
restless watch . --run "make test"
```

### Installation

```bash
make build
sudo make install
```

See the [main README](../README.md) and [documentation](README.md) for comprehensive guides.

## Community & Support

- 📖 [CLI Documentation](CLI.md)
- 🎓 [Tutorial](zero-to-hero.md)
- 🔗 [Release Notes](RELEASE.md)
- 🐛 [Issue Tracker](https://github.com/bspippi1337/restless/issues)

---

**Restless** – Explore systems safely. Automate intelligently. Think Unix.
