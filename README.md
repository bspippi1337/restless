# Restless

<img src="assets/brand/restless-hero.svg" width="920" />

**Domain-first API discovery and interaction engine**  
Evidence-driven CLI/TUI (Beta track).

## Quickstart

```bash
restless discover openai.com --verify --fuzz --budget-seconds 20 --budget-pages 8 --save-profile openai
```

## Dynamic help

```bash
restless discover --help
```

## Profiles

Saved to `~/.config/restless/profiles/<name>.yaml` (Linux/macOS/Termux).


## Console (interactive)

Build and run requests interactively, then save them as snippets:

```bash
restless console --profile openai
```

Docs:
- `docs/CONSOLE.md`
- `docs/SNIPPETS.md`


## Website

GitHub Pages is deployed automatically from `docs/`.

- Repository: https://github.com/bspippi1337/restless
- Releases: https://github.com/bspippi1337/restless/releases
