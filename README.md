# restless

Reactive API discovery and Unix automation runtime.

Restless helps developers understand unfamiliar systems quickly.

It maps structure before extensive manual exploration is required.

## Features

- safe API discovery
- topology mapping
- filesystem-triggered command execution
- structural drift inspection
- shell-friendly output
- terminal-native workflows

## Example

Map a repository surface:

    restless map .

Safely inspect a remote API:

    restless scan https://example-api.dev

Watch a directory and react to changes:

    restless watch . --run "make test"

## Safe by default

Remote discovery uses only:

- GET
- HEAD
- OPTIONS

No mutation is performed during discovery.

## Philosophy

Restless follows traditional Unix design principles:

- composable tools
- inspectable behavior
- shell interoperability
- local-first execution
- minimal runtime assumptions

The project intentionally avoids heavyweight orchestration layers.

## Installation

Build locally:

    make build

Install:

    sudo make install

## Packaging goals

Restless aims to remain:

- Debian-friendly
- reproducibly buildable
- dependency-light
- terminal-native
- suitable for headless systems

## Project structure

Primary runtime code:

    cmd/
    internal/

Documentation:

    docs/

Historical experiments and prototypes:

    archive/

Old patch scripts and one-off migrations:

    archive/historical/

The live repository cleanup and verification tools are native C and Go:

    make garden          # audit and dry-run preview, no deletions
    make garden-check    # gofmt, vet, tests, and build
    make garden-apply    # explicitly remove only generated directories

See [Repository Garden](docs/repo-garden.md) for archival paths, safety limits, and cross-platform build notes.

## License

MIT

## Status

![CI](https://github.com/bspippi1337/restless/actions/workflows/ci.yml/badge.svg)
![Installer](https://github.com/bspippi1337/restless/actions/workflows/installer.yml/badge.svg)
