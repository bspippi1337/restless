# Restless Archive Policy

The root module of Restless is intentionally kept small, composable, and Unix-oriented.

Experimental GUI code, abandoned prototypes, and legacy runtime experiments are preserved
under `archive/` for historical reference only.

These archived components are NOT part of the supported runtime surface.

## Design principles

Restless prefers:

- terminal-native workflows
- composable command pipelines
- inspectable runtime behavior
- minimal dependency surfaces
- portable builds
- Debian-friendly packaging
- GNU/Linux interoperability

## Root module policy

The root `go.mod` should avoid:

- GUI frameworks
- OpenGL bindings
- desktop-specific runtime assumptions
- heavyweight orchestration stacks

GUI experiments belong in isolated modules under `archive/`.
