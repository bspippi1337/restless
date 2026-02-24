# Combos (Restless + unix tools + friends)

1) Restless + jq (format/filter)
```bash
restless probe https://api.example.com | jq .
```

2) Restless + tee + bat (log + pretty)
```bash
restless simulate https://api.example.com | tee /tmp/restless.json | bat --language json
```

3) Restless + fzf (pick something fast)
```bash
# if you have a command that lists endpoints, adapt the left side:
restless endpoints --profile example | fzf --height=40% --reverse
```

4) Restless + parallel (many targets quickly)
```bash
printf "%s\n" https://api.a.com https://api.b.com https://api.c.com \
  | parallel -j6 "restless probe {} > out{#}.json"
```

5) Restless + entr (rerun on file change)
```bash
ls payload.json | entr -r restless smart https://api.example.com
```

6) Restless + httpx (subdomains → discover)
```bash
echo example.com | httpx -silent | xargs -I{} restless discover {} --verify
```

7) Restless + diff (drift detector)
```bash
restless discover example.com --verify --out /tmp/new.json
diff -u docs/known/example.json /tmp/new.json || true
```

8) Restless + rg (search evidence/output)
```bash
rg -n "openapi|swagger|/v1/|/api/" -S .
```
