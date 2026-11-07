# RESTLESS 42 Repository Garden

## What changed

The root has been reduced to project entrypoints, build metadata, security/community policy and package-manager integration. 53 previous files were **relocated**, not rewritten or stripped from Git history:

- Old root design notes and migrations: `archive/historical/root-docs/`
- A former temporary APK handoff and its checksum: `archive/historical/root-docs/FLYVEFAEN_TRANSFER.md`
- One-off root patchers: `archive/historical/root-scripts/`
- Old patch/release/repair experiments from `scripts/`: `archive/historical/one-off-scripts/`
- Retired integration workflows: `archive/historical/workflows/`
- An unused root web prototype: `archive/historical/root-site/index.html`
- Empty root marker files: `archive/historical/root-artifacts/`
- Reproducible-build notes, which remain relevant: `docs/builds/reproducible.md`

Active build, package, release and signing configuration remain in place. Notably `README.md`, `LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`, `Makefile`, `go.mod`, `go.sum`, `debian/`, `Dockerfile` and the current `install.sh` are intentionally kept at their familiar locations.

## Shell-to-native migrations

`scripts/cleanup.sh` -> `tools/c/restless-clean.c` (C11). The C cleaner supports POSIX and Windows native filesystem APIs. It does **not** invoke a shell, execute git, rewrite code, modify metadata, or access the network.

- Default mode is **dry-run**: prints only generated directories it would remove.
- `--apply` explicitly enables cleanup of `build/`, `dist/`, `bin/`, `logs/` and `.fixall-logs/` only.
- Requires a repository root containing `go.mod` and `.git`.
- Preflights all cleanup trees; refuses symlinks, junctions, reparse points and special files; limits recursion.
- Does not recursively delete arbitrary `*.log` throughout the working tree. This is intentionally narrower than the legacy shell script.

Example on Debian/WSL:

```sh
cc -std=c11 -D_DEFAULT_SOURCE -O2 -Wall -Wextra -Werror \
  -o .restless/restless-clean tools/c/restless-clean.c
.restless/restless-clean --dry-run
# Only after reviewing the preview:
.restless/restless-clean --apply
```

`scripts/doctor.sh` -> `cmd/restless-doctor/` (pure Go standard library). It checks `go.mod`, formatting, `go vet`, tests and build, logging steps and honoring a timeout. It never executes commands through a shell and does not run `go mod tidy`, force-formats, push commits or delete source files. The compiled test binary is placed in a temporary directory and discarded.

```sh
go run ./cmd/restless-doctor --audit-only
go run ./cmd/restless-doctor --timeout=10m
```

Run `make garden` for a read-only root audit and C cleaner preview, `make garden-check` for doctor verification, or `make garden-apply` for explicit cleanup.

## Continuous maintenance

`.github/workflows/repo-garden.yml` runs weekly and on pull requests, plus manual triggers. It checks root layout, Go doctor tests, builds the C cleaner on Linux/macOS, and tests dry-run, apply, symlink refusal, and source preservation. Windows runs the Go doctor checks. C compilation on Windows is supported in source via Windows APIs but is **not independently CI-verified yet**.

This job reports problems; it **does not automatically delete files or push changes to main**.

## Further translation candidates

The remaining shell scripts fall into three groups. One-off patchers belong in the historical archive; CI/packaging glue may remain shell until its real dependencies are evaluated; maintained operational helpers may be migrated to native Go or C case by case.

Do not naively convert a Debian packaging process to C: `dpkg-buildpackage`, signing tools, platform package managers and GitHub Actions are external build-system interfaces. Native code can orchestrate them safely, but cannot remove the host-specific package tooling.

Git history is preserved. If anything historical is needed, inspect `archive/historical/` or use `git log --follow -- <new-path>`.
