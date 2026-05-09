# Restless Output Contract

Restless output should remain:

- stable
- grep-friendly
- pipe-friendly
- JSON-compatible
- human-readable

## Human output

Human-oriented output may contain:

- ANSI colors
- compact summaries
- progress hints

Human output should still remain concise and line-oriented.

## JSON output

JSON output is considered machine-facing.

JSON mode should:

- emit valid JSON only
- avoid ANSI escape sequences
- avoid mixed log formats
- remain backwards compatible whenever possible

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | success |
| 1 | runtime failure |
| 2 | usage/configuration error |

## Philosophy

Restless prefers predictable text streams over opaque dashboards.
