# ⚡ Restless

**Discover the real structure of any API.**

Restless explores APIs the same way engineers do when documentation fails:  
by probing them.

It discovers endpoints, infers structure, detects schemas, and maps the topology of the API automatically.

Think of it as:

- **nmap** — but for APIs  
- **ripgrep** — but for endpoints  
- **graphviz** — but automatic  

```
scan → map → inspect
```

---

# Why Restless Exists

API documentation is often:

- incomplete
- outdated
- partially generated
- hiding internal endpoints
- inconsistent with reality

Restless assumes documentation may be wrong.

So it **discovers the API instead.**

---

# Terminal Demo

```
$ restless scan https://api.github.com

discovering endpoints...
probing routes...
inferring schema...

✔ 48 endpoints discovered
✔ pagination pattern detected
✔ authentication style inferred
✔ swagger detected

scan complete
```

Now map the API:

```
$ restless map
```

```
https://api.github.com
│
├── /users
│   ├── /{username}
│   │   ├── /repos
│   │   ├── /followers
│   │   └── /following
│
├── /repos
│   ├── /{owner}/{repo}
│   │   ├── /issues
│   │   ├── /pulls
│   │   └── /actions
│
└── /search
    ├── /repositories
    └── /issues
```

Inspect a single endpoint:

```
$ restless inspect /repos/{owner}/{repo}
```

```
METHODS
  GET

PARAMETERS
  owner   string
  repo    string

RETURNS
  repository object
```

---

# ASCII API Topology

Restless renders API structure directly in the terminal.

```
$ restless map --ascii
```

```
api.company.com
│
├─ auth
│  ├─ login
│  └─ refresh
│
├─ users
│  ├─ list
│  └─ profile
│
└─ billing
   ├─ invoices
   └─ payments
```

Large APIs become understandable **instantly**.

---

# Swagger / OpenAPI Detection

Restless automatically probes for schema definitions:

```
/swagger.json
/openapi.json
/api-docs
/v1/swagger
/docs/openapi
```

Example:

```
$ restless scan https://service.internal

probing schema endpoints...

✔ openapi detected: /openapi.json
✔ merging schema with discovered routes
```

This allows:

- schema inspection
- endpoint validation
- documentation drift detection

---

# Killer Use Cases

## Reverse engineer an unknown API

```
restless scan https://api.example.com
restless map
```

Understand the API structure instantly.

---

## Discover undocumented endpoints

```
restless scan https://internal.company.com
```

Find routes that never made it into documentation.

---

## Explore microservice gateways

```
restless scan http://gateway.local
```

Reveal services hidden behind routing layers.

---

## Compare staging vs production

```
restless scan https://staging.api
restless scan https://prod.api
```

Detect API drift before deployments break clients.

---

## Generate instant API documentation

```
restless map > api-topology.txt
```

Drop it into Slack or a ticket.

---

# Installation

Build from source:

```
git clone https://github.com/bspippi1337/restless
cd restless
make build
```

Binary appears in:

```
build/restless
```

Install system-wide:

```
sudo make install
```

---

# Quickstart

```
restless scan https://api.github.com
restless map
restless inspect /users/{username}
```

---

# Commands

```
restless scan <url>      discover API endpoints
restless map             generate endpoint topology
restless inspect         inspect endpoint details
```

---

# Architecture

```
cmd/restless
   │
   ├── internal/cli
   ├── internal/core
   │      ├── scanner
   │      ├── mapper
   │      └── inspector
   │
   └── internal/ui
```

Pipeline:

```
target API
   ↓
endpoint discovery
   ↓
structure inference
   ↓
topology generation
```

---

# Design Philosophy

Restless follows a few simple rules:

- **fast**
- **single binary**
- **terminal first**
- **discover reality**

Documentation can lie.

APIs don't.

---

# Roadmap

Planned features:

- SVG topology export
- Graphviz integration
- API diff engine
- authentication plugins
- fuzzing mode
- request replay

---

# Contributing

```
make test
make build
```

Pull requests welcome.

---

# Author

Pippi Tednes

---

# License

MIT
