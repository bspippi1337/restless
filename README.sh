# Restless

**Terminal-first API workbench. Deterministic. Scriptable. Fast.**

Restless turns your terminal into a focused API command center.

Probe. Inspect. Snapshot. Diff. Repeat.

---

## 📌 Current Milestone

### v5.0.0 — Core Reset

This release marks a structural reset of the project.

- GUI layer removed from root module
- Fyne dependency eliminated from core
- Repository history cleaned (large binaries removed)
- CI stabilized to core-only
- Clear separation between core and experiments

Restless is now lean, fast to clone, and terminal-native by design.

---

## Why Restless?

APIs are infrastructure.

Most tools around APIs drift toward:

- Browser tabs
- Heavy UIs
- Fragile state
- Manual workflows

Restless goes the opposite direction.

Built for:

- Developers who live in the terminal
- CI pipelines that must stay deterministic
- Repeatable inspection
- Script-first workflows
- Minimal surface area

No GUI coupling.  
No hidden state.  
No ceremony.

Just signal.

---

## Install

```bash
git clone https://github.com/bspippi1337/restless.git
cd restless
go build -o restless ./cmd/restless
```
