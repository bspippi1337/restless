## OpenAPI suggestions (adaptive, but contract stays the truth)

After running Restless against a host, it stores an observational snapshot locally and can generate spec improvement suggestions:

```bash
restless smart https://api.example.com

# generate suggestions (requires previous drift observations)
restless openapi suggest https://api.example.com --out suggestions --min-count 3
```

Outputs:
- `suggestions/<host>-<timestamp>.md`
- `suggestions/<host>-<timestamp>.json`
- `suggestions/<host>-<timestamp>.plan.json`
