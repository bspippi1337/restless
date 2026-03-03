# restless ⚡

**Restless** is a fast CLI for exploring and validating REST APIs from their OpenAPI / Swagger specs.

It can:

- discover OpenAPI specs from a target
- probe endpoints automatically
- generate API topology maps
- export diagrams (SVG)
- run fast verification of APIs

Think of it as a small **API reconnaissance + validation toolkit**.

---

## Install

Clone and run directly with Go:

go run ./cmd/restless <command>

or build the binary:

go build -o restless ./cmd/restless

---

## Commands

scan a target for OpenAPI specs

restless scan https://target

verify an API from a spec

restless verify --spec openapi.yaml

verify directly from remote spec

restless verify --spec https://petstore.swagger.io/v2/swagger.json

map API topology

restless map openapi.yaml

generate SVG diagram

restless map openapi.yaml --svg > api-map.svg

run full pipeline (discover + verify)

restless inspect https://petstore.swagger.io

---

## Example

restless inspect https://petstore.swagger.io

Typical output:

OK    GET    /pet/1
OK    GET    /store/inventory
WARN  POST   /pet

Summary:
OK   5
WARN 14
FAIL 0

---

## API topology

Generate a diagram:

restless map https://petstore.swagger.io/v2/swagger.json --svg > api.svg

Embed it in docs:

![API Map](docs/demo.svg)

---

## Features

OpenAPI / Swagger discovery  
automatic path parameter generation  
HTTP probing with latency measurement  
ASCII API topology map  
SVG API diagrams  
CLI-first workflow  

---

## Project layout

restless
├─ cmd/restless        CLI
├─ internal/core       core data types
├─ internal/httpx      HTTP executor
├─ internal/openapi    spec loader
├─ internal/probe      request generation
├─ internal/graph      API map builder
├─ internal/discovery  swagger detection
├─ internal/report     output formatting
├─ examples            sample specs
└─ docs                generated diagrams

---

## Development

format code

go fmt ./...

run tests

go test ./...

build binary

go build ./cmd/restless

---

## License

MIT
