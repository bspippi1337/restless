# Architecture

Restless is structured around modular execution layers:

cmd/restless        → CLI router
internal/core/app   → module registry
internal/modules/
    openapi         → spec engine
    session         → templating
    bench           → load testing
    export          → artifacts
internal/version    → version injection

Design principles:

- Deterministic execution
- CI-friendly
- Static builds
- Zero external runtime dependencies
