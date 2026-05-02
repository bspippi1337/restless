# Repository hygiene

The repository root should contain source, packaging, documentation, and project metadata only.

Generated files belong outside git, especially:

- scan output
- `.svg`, `.dot`, and `.svg.dot` API maps
- temporary JSON/HAR files
- local test reports
- one-off patch scripts
- local stress-test logs

Use `examples/`, `docs/`, `assets/`, `scripts/`, `tools/`, `cmd/`, `internal/`, and `pkg/` for intentional project files.
