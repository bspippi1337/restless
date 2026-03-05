# RESTLESS — Zero to Hero

This guide takes you from:

- Clean machine
- No API knowledge
- No tooling
- No structure

To:

- Structured API exploration
- Reproducible debugging
- CI validation
- Production-grade workflow

We will simulate a real SaaS backend called:

    https://api.example.com

---

# 0. Install

```bash
go install github.com/bspippi1337/restless/cmd/restless@latest
```

Verify:

```bash
restless --version
```

---

# 1. You Know Nothing About the API

You join a team.
They give you only a base URL.

No docs.
No Postman collection.

## Step 1: Surface Discovery

```bash
restless probe https://api.example.com
```

Restless maps:

- Available routes
- Status responses
- Surface patterns
- Response structure hints

This replaces random curl guessing.

---

# 2. Import Official Spec (If Available)

You later receive:

    openapi.yaml

Import it:

```bash
restless openapi import ./openapi.yaml
```

Now you have structured endpoint access.

List discovered endpoints:

```bash
restless list
```

---

# 3. Create Structured Environment Sessions

You have:

- staging
- production

Create sessions:

```bash
restless session create staging --base https://staging.api.example.com
restless session create prod --base https://api.example.com
```

Add auth:

```bash
restless session set staging --header "Authorization: Bearer $STAGING_TOKEN"
restless session set prod --header "Authorization: Bearer $PROD_TOKEN"
```

Switch instantly:

```bash
restless session load staging
```

You now have clean environment isolation.

---

# 4. Execute Real Workflows

Example: Create a user.

```bash
restless run POST /users \
  --body '{"name":"Alice","role":"admin"}'
```

Fetch the user:

```bash
restless run GET /users/42
```

Pipe to jq:

```bash
restless run GET /users/42 | jq '.email'
```

Structured + composable.

---

# 5. Debug a Production Issue

Users report:

"Order API fails intermittently."

Switch session:

```bash
restless session load prod
```

Reproduce:

```bash
restless run GET /orders/17291
```

Save result:

```bash
restless run GET /orders/17291 > order.json
```

Add debug header:

```bash
restless run GET /orders/17291 \
  --header "X-Debug: true" > order-debug.json
```

Compare:

```bash
diff -u order.json order-debug.json
```

Root cause identified.

---

# 6. Export Evidence

Generate Markdown report:

```bash
restless export --format md --out reports/order-17291
```

Generate JSON artifact for CI:

```bash
restless export --format json --out artifacts/
```

No screenshots.
No manual copying.

---

# 7. Validate API Drift in CI

Add to your CI pipeline:

```yaml
- name: Validate API
  run: |
    restless openapi import ./openapi.yaml
    restless probe $BASE_URL
    restless export --format json --out surface/
    diff -r openapi.yaml surface/
```

Fail build if API changed unexpectedly.

Signal corrected before deployment.

---

# 8. Load Testing Quick Sanity Check

Use GNU parallel:

```bash
seq 1 100 | parallel -n0 restless run GET /health
```

Count status codes:

```bash
... | jq -r '.status' | sort | uniq -c
```

Identify rate limits instantly.

---

# 9. Makefile Integration

```make
probe:
	restless probe $(BASE_URL)

validate:
	restless openapi import ./openapi.yaml
	restless probe $(BASE_URL)
	restless export --format json --out surface/
	diff -r openapi.yaml surface/
```

Run:

```bash
make validate
```

Now your API validation is one command.

---

# 10. Docker Usage

```Dockerfile
FROM golang:1.22
RUN go install github.com/bspippi1337/restless/cmd/restless@latest
ENTRYPOINT ["restless"]
```

Use in CI containers without local install.

---

# 11. Full Production Workflow Example

```bash
restless session load prod
restless run GET /health
restless run GET /users/42
restless run POST /audit \
  --body '{"action":"debug","user":"42"}'
restless export --format md --out audit-session/
```

Everything reproducible.
Everything structured.

---

# What You Now Have

You went from:

- No API knowledge
- No structure
- Manual curl chaos

To:

- Structured exploration
- Repeatable debugging
- CI enforcement
- Exportable evidence
- Environment isolation

That is Zero → Hero.

---

# Philosophy

Restless does not replace Unix tools.

It makes them precise.

It does not create noise.

It corrects signal.
