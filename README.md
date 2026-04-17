# RESTLESS

<p align="center">
<b>Terminal-native API topology discovery engine</b>
</p>

Restless discovers unknown API surfaces, infers structure, and renders a usable topology model directly from a target URL.

---

## QUICK START

```bash
restless https://api.github.com
```

Example output:

```
probing API surface
inferring structure
building topology
rendering graph

Detected: REST + GraphQL
Endpoints: 27
Docs: /openapi.json
Topology depth: 3

Graph written: api.github.com.svg
```

Result:

```
api.github.com.svg
```

---

## WHAT RESTLESS DOES

Pipeline:

```
target
  ↓
discovery
  ↓
probing
  ↓
inference
  ↓
topology
  ↓
graph
```

Instead of calling endpoints manually, Restless builds a structural model of an API automatically.

---

## CORE COMMANDS

```
restless discover
restless map
restless inspect
restless call
restless learn
restless graph
restless completion
```

Experimental:

```
restless blckswan
restless swarm
restless smart
```

---

## INSTALL

```
curl -sSL https://bspippi1337.github.io/restless/install.sh | sh
```

or

```
wget -qO- https://bspippi1337.github.io/restless/install.sh | sh
```

---

## OUTPUT

Restless produces:

- endpoint discovery
- documentation detection
- topology inference
- graph visualization

Example topology fragment:

```
/users
├── /users/{id}
│   ├── /repos
│   └── /followers
└── /orgs/{org}
```

---

## LICENSE

MIT
