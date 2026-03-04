# Restless

**Deterministic reduction of complex system state**  
Minimal repro + explain for failing API calls, JSON payloads, config drifts and more.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Latest Release](https://img.shields.io/github/v/release/bspippi1337/restless?color=green)](https://github.com/bspippi1337/restless/releases)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen) <!-- oppdater med ekte badge når CI er oppe -->

**Restless** hjelper deg å finne **nøyaktig** hva som gjør at en API-request feiler – uten flaks, uten "det funket på min maskin".  
Den tar stor, kompleks input (JSON, headers, env, OpenAPI spec) og reduserer den til det **minste mulige settet** som fortsatt reproducerer feilen – deterministisk og reproduserbart.

```bash
# Typisk bruk i pipeline
curl -s https://api.example.com/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -d @huge-payload.json \
  | jq . \
  | restless explain --spec openapi.yaml --live
