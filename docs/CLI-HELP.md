# restless v2 · Help

Restless is a **CLI-first** API client with a polished TUI and a discovery engine that can start from **one input**: a domain.

## Quick start (30 seconds)
1) Run:
```bash
restless
```
2) In **Connect & Discover**, enter:
```text
bankid.no
```
3) Press **Ctrl+D** to discover.
4) Switch tab to **Request Builder** to see suggestions.

## Keys
- `Ctrl+D` discover (wizard)
- `Tab` / `Shift+Tab` switch tabs
- `?` open help tab
- `q` quit

## Commands
### Help
```bash
restless --help
restless help
```

### Discover from a domain (CLI)
```bash
restless discover bankid.no
restless discover openai.com --json
```

### Clean & diagnose
```bash
restless doctor
```

## What "Fuzzer Mode" means here
- Parses public documentation first (OpenAPI / docs pages)
- Scrapes hints conservatively
- Expands endpoints from *seeds*, then verifies safely
- No brute force, no auth bypass, no write calls by default

## Tips
- If discovery finds **401/403**, that still counts as a valid endpoint (auth required).
- Prefer sandbox domains when available (developer portals often link them).
