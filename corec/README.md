# Restless C Core (skeleton)

This folder introduces a **dependency-minimal C core** for Restless.

Goal:
- Keep a hardened core as a **single native binary** (`restless-core`)
- Avoid vendored third-party libs (no giant dependency trees)
- Use OS TLS backends when possible (WinHTTP / SecureTransport / system OpenSSL) rather than "roll your own TLS"

Status: **skeleton** (v0) — compiles and runs `doctor` + stub `discover`.

## Build (Linux/Debian/Termux)

```sh
cd corec
make
./bin/restless-core doctor
./bin/restless-core discover openai.com
```

## Notes on HTTPS

A fully correct TLS implementation should NOT be written from scratch.
This core uses a backend abstraction. In the skeleton, HTTPS is stubbed.

Next steps:
- Windows: implement HTTPS via WinHTTP (no external libs)
- macOS: implement via SecureTransport/Network.framework
- Linux/Termux: optionally link system OpenSSL (installed via package manager), or provide HTTP-only fallback
