# QA Acceptance Tests (v2 One-Click / Safe-by-Default)

These are human-run acceptance tests to validate Restless behaves as a safe, intuitive product.

## Environments
- Debian/Ubuntu (incl. WSL)
- Windows
- macOS
- Optional: flaky network

## 1) First-run UX
1. Launch `restless`.
2. Wizard shows a single domain input, focused.
3. Hints visible: Tab, Ctrl+D, ?, q.

Pass: no crash, no config prompts, no stack traces.

## 2) Domain-only discovery
Try:
- `example.com` (graceful)
- `openai.com` or known API host

Steps: Enter domain → Ctrl+D → observe Base URL + endpoints or graceful “no results”.

Pass: returns to idle; no error wall.

## 3) Safety defaults
Pass: no POST/PUT/PATCH/DELETE issued automatically; verification uses GET/HEAD/OPTIONS only.

## 4) 401/403 handling
Pass: 401/403 counts as “exists (auth required)” and is included.

## 5) Doctor
Create: `bin/ dist/ build/ logs/` + `foo.log`. Run `restless doctor`.

Pass: removes only known targets; readable report; does not touch secrets/keys.

## 6) Help
Press `?`, scroll, switch tabs.

Pass: renders even if docs missing; no corruption.

## 7) Repeatability
Run discovery + doctor repeatedly.

Pass: stable; bounded runtime; no destructive behavior outside cleanup targets.
