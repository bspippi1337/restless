# RESTLESS

<p align="center">
<b>API topology inference engine</b><br>
Discover • Model • Inspect • Explain any API surface
</p>

---

## INSTALL

```bash
curl -sSL https://bspippi1337.github.io/restless/install.sh | sh
```

eller

```bash
wget -qO- https://bspippi1337.github.io/restless/install.sh | sh
```

---

## QUICK START

```bash
restless scan https://api.github.com
restless map
restless inspect GET /users
restless graph
restless teach
```

Restless lagrer automatisk siste analyse som session‑state.
Du trenger normalt ikke oppgi target mer enn én gang.

---

## WHAT RESTLESS IS

Restless er en API‑rekonstruksjonsmotor.

Den:

• finner endpoints
• analyserer struktur
• identifiserer relasjoner
• utleder schema
• bygger topologi
• foreslår neste steg

Dette gjør Restless til en "Copilot for APIs" i terminalen.

---

## CORE WORKFLOW

```text
scan → learn → map → inspect → graph → teach → copilot
```

Beskrivelse:

| command | role |
|--------|------|
| scan | rask endpoint discovery |
| learn | dokumentasjonsdrevet analyse |
| map | strukturell oversikt |
| inspect | endpoint‑analyse |
| graph | visualiser API‑topologi |
| teach | forklar API‑struktur |
| copilot | foreslå neste steg |

---

## EXAMPLE SESSION

```bash
restless scan api.example.com
restless map
```

Output:

```text
/
├── users
├── users/{id}
│   └── posts
└── health
```

Deretter:

```bash
restless inspect GET /users
```

Gir:

```text
FIELDS
id:number
name:string
email:string
```

---

## DISCOVERY ENGINE

Restless bruker en bounded crawler som:

• følger kun samme host
• detekterer JSON‑schema
• identifiserer nested endpoints
• stopper deterministisk
• respekterer depth‑grenser

Dette gjør discovery trygg og rask.

---

## GRAPH OUTPUT

```bash
restless graph
```

Genererer:

```text
api.svg
```

eller DOT‑output:

```bash
restless graph target --format dot
```

---

## SESSION STATE

Restless lagrer automatisk:

```text
~/.restless_state.json
```

Dette gjør CLI‑workflow interaktiv:

```bash
scan
map
inspect
teach
```

uten å oppgi target på nytt.

---

## INTELLIGENCE FEATURES

Restless oppdager automatisk:

• OpenAPI
• Swagger
• GraphQL
• health endpoints
• metrics endpoints
• nested resource trees
• collection patterns

---

## ARCHITECTURE

```text
CLI
 ├ discovery engine
 ├ topology inference
 ├ schema detection
 ├ fuzz heuristics
 └ session state
```

Motoren er designet for komponerbar API‑rekonstruksjon.

---

## PROJECT STRUCTURE

```text
cmd/restless
internal/cli
internal/discovery
internal/core
internal/state
internal/tui
```

Legacy‑kode ligger isolert i archive/.

---

## DESIGN PRINCIPLES

Discovery first

Graph‑based thinking

Composable engines

Session‑aware workflow

---

## LICENSE

MIT
