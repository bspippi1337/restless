# RESTLESS

<p align="center">
<svg width="300" height="300" viewBox="0 0 400 400" xmlns="http://www.w3.org/2000/svg">
<defs>
<linearGradient id="bolt" x1="0%" y1="0%" x2="100%" y2="100%">
<stop offset="0%" stop-color="#00eaff"/>
<stop offset="100%" stop-color="#8a2be2"/>
</linearGradient>
<filter id="glow">
<feGaussianBlur stdDeviation="6" result="blur"/>
<feMerge>
<feMergeNode in="blur"/>
<feMergeNode in="SourceGraphic"/>
</feMerge>
</filter>
</defs>
<circle cx="200" cy="200" r="170" fill="#0b0f19" stroke="#222" stroke-width="4"/>
<path d="M210 70 L160 210 L210 210 L170 330 L260 170 L210 170 Z"
fill="url(#bolt)" filter="url(#glow)"/>
</svg>
</p>

<p align="center">
<b>API reconnaissance framework</b><br>
Discover • Map • Probe • Understand any API surface
</p>

---

INSTALL

curl -sSL https://bspippi1337.github.io/restless/install.sh | sh

eller

wget -qO- https://bspippi1337.github.io/restless/install.sh | sh

---

FIRST RUN

restless blckswan https://api.github.com

---

WHAT RESTLESS DOES

Restless automatically explores an API surface and builds a structural understanding of it.

Instead of manually browsing endpoints one by one, Restless performs automated reconnaissance.

Workflow:

target
 │
 ▼
endpoint discovery
 │
 ▼
probing & fuzzing
 │
 ▼
documentation detection
 │
 ▼
topology inference
 │
 ▼
API insight

---

EXAMPLE

Routes discovered: 27

/users
├── /users/{id}
│   ├── /repos
│   └── /followers
└── /orgs/{org}

---

CORE COMMANDS

restless discover
restless inspect
restless scan
restless map

restless swarm
restless magiswarm
restless octoswan

restless auto
restless blckswan
restless smart

---

ARCHITECTURE

CLI
 │
 ├─ application layer
 │
 ├─ core engines
 │   ├─ discovery
 │   ├─ probing
 │   ├─ topology
 │   ├─ swarm engines
 │   └─ fuzzing
 │
 ├─ modules
 │   ├─ OpenAPI intelligence
 │   ├─ export
 │   └─ sessions
 │
 └─ infrastructure
     ├─ HTTP engine
     ├─ persistence
     ├─ logging
     └─ terminal UI

---

PHILOSOPHY

Discovery first  
Assume nothing about an API and learn its structure dynamically.

Graph thinking  
Treat APIs as connected systems rather than isolated endpoints.

Composable engines  
Small probing engines combine into powerful reconnaissance pipelines.

---

PROJECT STRUCTURE

cmd/restless
internal/core
internal/modules
internal/swarm
internal/topology
docs/
tools/

---

LICENSE

MIT
