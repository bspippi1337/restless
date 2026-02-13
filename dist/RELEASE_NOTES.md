# Restless v0.3.0-3-g7f77255 (Termux build)

This release packages a **pure-Go** Restless binary for Android/Termux with a minimal, investor-friendly distribution layout.

## What you get
- **`restless`**: single-file CLI/TUI binary (no config-file required to start)
- **Connect & Discover** workflow: discover API entry points from **domain-only input**
- **Fuzzer mode** (scaffolding): doc-driven + light scraping style fuzzing
- **Doctor**: environment + repo sanity checks (designed to self-heal common build footguns)
- **SSE streaming** support (where configured)

## Installation (Termux)
```sh
unzip restless_v0.3.0-3-g7f77255_android_termux_aarch64.zip
cd release
chmod +x restless
./restless --help
```

## Quick examples
### Discover from domain-only
```sh
./restless discover bankid.no
./restless discover openai.com
```

### Run the TUI
```sh
./restless tui
```

### Validate environment
```sh
./restless doctor
```

## Files in this release
- `restless` (binary)
- `README.md` + `LICENSE` (if present)
- `docs/QUICKSTART.md`
- `docs/CLI-HELP.md` (if present)
- `examples/demo.sh`

## Notes
- Built with: `CGO_ENABLED=0`, tags: `netgo osusergo`, stripped with `-ldflags \"-s -w\"`.
- This is an **alpha** distribution intended for fast iteration and investor demos.
