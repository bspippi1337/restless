# Restless ⚡

<p align="center">
<b>Universal API Client — built for the terminal.</b><br/>
Discover • Probe • Simulate • Export
</p>

<p align="center">
<img src="assets/banner-github.svg" alt="Restless Banner"/>
</p>

---

## What is Restless?

Restless is a universal API client designed for engineers who prefer clarity, speed, and control.

It provides:

- Fast endpoint discovery
- Interactive request simulation
- Smart guided workflows
- Clean, scriptable output
- Minimal dependency footprint

Restless lives where you live: the terminal.

---

## Installation

### Debian / Ubuntu

```bash
echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless
```

### From Source

```bash
go install github.com/bspippi1337/restless/cmd/restless@latest
```

---

## Quick Start

```bash
restless probe https://api.github.com
restless simulate https://api.github.com
restless smart https://api.github.com
```

---

## Core Philosophy

- Terminal-first
- Precision over noise
- Scalable architecture
- Composable workflows
- Production-ready structure

---

## Architecture Overview

```
CLI
  ↓
Command Layer (probe / simulate / smart)
  ↓
Discovery Engine
  ↓
Core Engine
  ↓
Simulator / Export
  ↓
HTTP Client
```

---

## Roadmap

- OpenAPI auto-detection
- Intelligent autocomplete
- Enhanced TUI mode
- Plugin-style extensibility

---

MIT Licensed.
