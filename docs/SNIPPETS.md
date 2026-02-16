# Snippets (v1)

Snippets are small, reusable request recipes saved per profile, with lightweight stats (useCount / success / latency).

Default location:
`~/.config/restless/snippets/<profile>/<name>.yaml`

## Create snippets (recommended)

```bash
restless console --profile openai
# suggest -> run -> save models
```

## List

```bash
restless snippets list --profile openai
```

## Run

```bash
restless snippets run --profile openai models
```

## Export (copy/paste)

```bash
restless snippets export --profile openai models --format curl
restless snippets export --profile openai models --format httpie
```

## Pin/unpin

```bash
restless snippets pin --profile openai models
restless snippets unpin --profile openai models
```

## Stats

Each snippet file tracks:

- useCount
- successCount / failCount
- avgLatencyMs

A run history is appended to:
`~/.config/restless/history/<profile>.jsonl`
