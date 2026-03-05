# Demo (15–20 seconds)

Goal: show *discovery + structure + state*.

## Recommended sequence

```bash
restless probe https://httpbin.org
restless list
restless run GET /headers
restless session
```

## Record with asciinema

```bash
asciinema rec demo.cast
# run the sequence above
exit
```

Upload to asciinema.org, then add the link near the top of README.

## Record a GIF (optional)

If you prefer a GIF for README:

- terminalizer (simple)
- vhs (pretty)
- asciinema + agg (clean)

Keep it boring: no weird prompt, no distractions, just the tool.
