# Contributing to Restless

Thank you for contributing to Restless.

Restless is a transparent, deterministic, distribution-friendly API topology inference CLI. Contributions that improve correctness, portability, documentation quality, reproducibility, and packaging quality are especially welcome.

## Build from source

```bash
git clone https://github.com/bspippi1337/restless.git
cd restless
go build ./cmd/restless
```

To inject release metadata:

```bash
go build ./cmd/restless \
  -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=dev \
            -X github.com/bspippi1337/restless/internal/version.Commit=$(git rev-parse --short HEAD) \
            -X github.com/bspippi1337/restless/internal/version.Date=$(date -u +%Y-%m-%d)"
```

## Development guidelines

Please keep changes:

- deterministic
- documented
- dependency-light
- portable across Linux distributions
- bounded and explicit in network behaviour

Do not add telemetry, analytics, hidden background communication, generated binary blobs, or vendored executables.

## Testing

Prefer tests and examples that can run without network access.

Offline fixtures should live in:

```text
testdata/
```

Example fixture-oriented workflow:

```bash
restless inspect --fixture testdata/sample.json
```

## Command-line behaviour

CLI output should be stable and script-friendly. Sort paths, methods, and schema fields where practical.

Errors should be explicit and written to stderr by command execution paths.

## Commit style

Prefer small logical commits:

```text
cli: improve inspect help
state: migrate session state to XDG
man: add restless manual page
```

## Release process

A release should include:

1. a tagged version
2. updated documentation
3. generated or updated manual page content
4. reproducible version metadata via linker flags

## Distribution friendliness

Restless aims to remain suitable for packaging in distributions such as Debian.

Please avoid changes that require network access during build, proprietary tooling, or generated artifacts that cannot be reproduced from source.
