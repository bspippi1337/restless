# Restless v2 Architecture

## Goals
- Single binary.
- Modular codebase (compile-time modules).
- Small, stable core.
- Modules can evolve fast without breaking the core.

## Rules
- `internal/core` MUST NOT import `internal/modules`.
- `internal/modules` MAY import `internal/core`.
- UI layers (cli/tui/gui) depend on core + modules, never the other way around.

## Layout
- `internal/core/*`  : stable foundation (config, transport, engine, store, shared types)
- `internal/modules/*`: feature modules (openapi, sessions, bench, export, ui modules)
- `internal/ui/*`    : renderers and interaction layers
