# Restless

API reconnaissance framework for discovering, mapping and interrogating modern APIs.

Restless is a modular toolkit for exploring unknown APIs. It combines endpoint discovery, documentation scraping, OpenAPI intelligence, topology mapping and swarm-style probing into a single framework.

The goal is simple: point Restless at an API and quickly understand its structure, behavior and surface area.

---

## What Restless Does

Modern APIs are often poorly documented, partially exposed, or distributed across multiple protocols and endpoints.

Restless helps you answer questions like:

- What endpoints actually exist?
- Which parameters are accepted?
- Is there an OpenAPI spec hiding somewhere?
- How do endpoints relate to each other?
- What does the full topology of the API look like?

Restless performs automated reconnaissance to build a structured view of an API.

Typical workflow:

target → discovery → probing → documentation parsing → topology → report

---

## Capabilities

### Endpoint Discovery

Finds endpoints through probing, crawling and inference.

Examples:

- path probing
- documentation scraping
- heuristic discovery
- OpenAPI detection

---

### API Topology Mapping

Restless builds a structural graph of an API.

Example:

/users
│
├── /users/{id}
│   ├── /users/{id}/repos
│   └── /users/{id}/followers
│
└── /orgs/{org}

The topology engine identifies relationships between endpoints and resources.

---

### OpenAPI Intelligence

Restless attempts to locate and analyze API specifications.

Supported sources include:

- OpenAPI / Swagger
- embedded schemas
- inferred structures
- documentation hints

---

### Swarm Reconnaissance

Parallel probing engines explore an API surface quickly.

Engines include:

- swarm
- magiswarm
- octoswan

These distribute discovery and probing tasks across workers.

---

### Documentation Extraction

Restless extracts information from:

- HTML docs
- markdown
- API explorer pages
- embedded schemas

This helps reconstruct APIs even when official documentation is incomplete.

---

## Architecture

Restless is built as a layered framework.

CLI
│
├── Application Layer
│
├── Core Engines
│   ├── discovery
│   ├── probing
│   ├── topology
│   ├── swarm engines
│   └── fuzzing
│
├── Modules
│   ├── OpenAPI
│   ├── export
│   ├── session
│   └── benchmarking
│
└── Infrastructure
    ├── HTTP clients
    ├── state persistence
    ├── logging
    └── UI / terminal output

This architecture allows the engines and modules to evolve independently.

---

## Command Line Interface

Restless exposes the framework through a CLI.

restless discover
restless scan
restless inspect
restless map

restless swarm
restless magiswarm
restless octoswan

restless auto
restless blckswan
restless smart

---

## Examples

Discover endpoints:

restless discover https://api.github.com

Generate a topology map:

restless map https://api.github.com

Run autonomous reconnaissance:

restless auto https://api.github.com

Full reconnaissance pipeline:

restless blckswan https://api.github.com

---

## Example Output

Restless stores session state locally.

Saved scan → ~/.restless_state.json  
Routes discovered: 27

Example topology:

/users
├── /users/{id}
│   ├── /users/{id}/repos
│   └── /users/{id}/followers
└── /orgs/{org}

---

## Use Cases

Restless can be used for:

- API exploration
- reverse engineering undocumented APIs
- security research
- API auditing
- documentation reconstruction
- developer tooling

---

## Project Structure

cmd/restless  
CLI entrypoint

internal/core  
core engines

internal/modules  
optional modules and integrations

internal/httpx  
HTTP layer

internal/swarm  
parallel probing engines

internal/topology  
API graph construction

---

## Philosophy

Restless is designed around three principles:

Discovery first  
Assume nothing about the API and learn its structure dynamically.

Graph thinking  
Treat APIs as connected systems rather than isolated endpoints.

Composable engines  
Small components working together form powerful reconnaissance pipelines.

---

## Status

Restless is under active development.

The framework is evolving toward a full platform for API reconnaissance.

---

## License

See LICENSE file.
