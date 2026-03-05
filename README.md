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
