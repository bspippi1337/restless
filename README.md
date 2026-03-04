<p align="center">

<h1>⚡ restless</h1>

<b>Map APIs. Break assumptions. Learn the system.</b>

CLI for discovering and testing HTTP APIs the way engineers actually explore systems.

</p>

<p align="center">

<svg viewBox="0 0 900 260" xmlns="http://www.w3.org/2000/svg">
<style>
text{font-family:monospace}
.node{fill:#0b0f19;stroke:#00e0ff;stroke-width:2}
.label{fill:#00e0ff;font-size:14px}
.arrow{stroke:#00e0ff;stroke-width:2;marker-end:url(#arrow)}
.pulse{stroke-dasharray:8 6;animation:flow 2s linear infinite}
@keyframes flow{
0%{stroke-dashoffset:20}
100%{stroke-dashoffset:0}
}
</style>

<defs>
<marker id="arrow" viewBox="0 0 10 10" refX="10" refY="5"
markerWidth="6" markerHeight="6" orient="auto">
<path d="M0 0 L10 5 L0 10 z" fill="#00e0ff"/>
</marker>
</defs>

<rect class="node" x="40" y="100" width="160" height="60" rx="6"/>
<text class="label" x="95" y="135">API</text>

<rect class="node" x="330" y="40" width="200" height="60" rx="6"/>
<text class="label" x="360" y="75">probe</text>

<rect class="node" x="330" y="160" width="200" height="60" rx="6"/>
<text class="label" x="360" y="195">smart simulate</text>

<rect class="node" x="650" y="100" width="200" height="60" rx="6"/>
<text class="label" x="710" y="135">signals</text>

<line class="arrow pulse" x1="200" y1="130" x2="330" y2="70"/>
<line class="arrow pulse" x1="200" y1="130" x2="330" y2="190"/>
<line class="arrow pulse" x1="530" y1="70" x2="650" y2="130"/>
<line class="arrow pulse" x1="530" y1="190" x2="650" y2="130"/>
</svg>

</p>

---

# 30-second pitch

Most API tools assume you **already know the API**.

Restless assumes you **don't**.

Instead of guessing endpoints with curl until something works, Restless explores the API surface like a curious engineer:

- probes endpoints
- tests HTTP methods
- detects auth boundaries
- observes real behaviour

The result is **signals**, not noise.

---

# Installation

### Go

```
go install github.com/bspippi1337/restless/cmd/restless@latest
```

### Build from source

```
git clone https://github.com/bspippi1337/restless
cd restless
make build
```

---

# Usage

```
restless <domain-or-url>

restless probe <domain-or-url>

restless smart <domain-or-url>

restless simulate <domain-or-url>

restless <METHOD> <url>
```

Example:

```
restless probe https://api.example.com
```

---

# Killer workflow

### Discover API surface

```
restless probe https://api.example.com
```

Signals revealed:

- possible endpoints
- supported HTTP methods
- auth hints
- status behaviour

---

### Behavioural simulation

```
restless smart https://api.example.com
```

Smart mode explores the API using realistic requests and observes:

- rate limits
- schema stability
- auth cliffs
- inconsistent responses

---

### Direct surgical request

```
restless GET https://api.example.com/health
restless POST https://api.example.com/v1/login
```

---

# Brutal one-liners

Find auth boundaries

```
restless probe https://api.example.com | grep 401
```

Detect rate limiting

```
for i in $(seq 1 40); do restless GET https://api.example.com/ping; done
```

Post-deploy smoke test

```
restless smart https://api.example.com
```

Probe every endpoint hint

```
restless probe https://api.example.com | sort
```

---

# Example helper script

```
#!/usr/bin/env bash

target=$1

echo "=== probing api ==="
restless probe "$target"

echo
echo "=== behavioural simulation ==="
restless smart "$target"

echo
echo "=== health check ==="
restless GET "$target/health"
```

---

# What restless focuses on

Restless tries to provide:

fast signals  
minimal noise  
realistic behaviour  

It complements tools like:

curl  
httpie  
Postman  
k6  

by focusing specifically on **understanding unknown APIs quickly**.

---

# Status

Active development.

Focus areas:

- smarter probing heuristics
- signal extraction
- CLI stability
- packaging

---

# License

MIT
