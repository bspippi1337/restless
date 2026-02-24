# Workflows

1) Domain → profile → TUI
```bash
restless discover example.com --verify --budget-seconds 20 --budget-pages 8 --save-profile example
restless tui
```

2) Capture output to file
```bash
restless probe https://api.example.com > /tmp/restless.json
```

3) Pretty-print JSON
```bash
restless probe https://api.example.com | jq .
```

4) Proxy debug
```bash
HTTPS_PROXY=http://127.0.0.1:8080 restless tui
```
