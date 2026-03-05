# ⚡ Restless

> Discover the real structure of any API.

Restless explores APIs the same way engineers do when documentation fails:
by probing them.

It discovers endpoints, infers structure, detects schemas, and maps the topology of an API automatically.

Think:

• **nmap** — but for APIs  
• **ripgrep** — but for endpoints  
• **graphviz** — but automatic  

```
scan → map → inspect
```

---

# Demo

```
$ restless scan https://api.github.com

probing API...
discovering routes...
inferring structure...

✔ 48 endpoints discovered
✔ pagination pattern detected
✔ auth style inferred
✔ OpenAPI detected

scan complete
```

Map the structure:

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

Inspect endpoint details:

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

Large APIs become understandable instantly.

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

Perfect for:

• terminal exploration  
• quick documentation  
• architecture discussions  

---

# Swagger / OpenAPI Detection

Restless automatically probes common schema locations:

```
/swagger.json
/openapi.json
/api-docs
/v1/swagger
/docs/openapi
```

Example:

```
$ restless scan https://internal.service

probing schema endpoints...

✔ OpenAPI detected
✔ schema imported
✔ merged with discovered routes
```

This allows:

• schema inspection  
• endpoint validation  
• documentation drift detection  

---

# Why Restless Exists

API documentation is often:

• incomplete  
• outdated  
• partially generated  
• hiding internal endpoints  

Restless assumes documentation may be wrong.

So it **discovers the API instead**.

---

# Real-World Use Cases

## Reverse engineer an unknown API

```
restless scan https://api.example.com
restless map
```

Understand the entire structure in seconds.

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

Reveal services behind routing layers.

---

## Detect staging vs production drift

```
restless scan https://staging.api
restless scan https://prod.api
```

Identify structural differences.

---

## Generate instant API documentation

```
restless map > api-structure.txt
```

Share topology with teammates instantly.

---

# Installation

Build from source:

```
git clone https://github.com/bspippi1337/restless
cd restless
make build
```

Binary:

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

• **fast**  
• **single binary**  
• **terminal first**  
• **discover reality**  

Documentation can lie.

APIs don't.

---

# Roadmap

Planned features:

• SVG topology export  
• Graphviz integration  
• API diff engine  
• authentication plugins  
• fuzzing mode  
• request replay  

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
