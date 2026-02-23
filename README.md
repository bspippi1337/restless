# Restless

```
██████╗ ███████╗███████╗████████╗██╗     ███████╗███████╗███████╗
██╔══██╗██╔════╝██╔════╝╚══██╔══╝██║     ██╔════╝██╔════╝██╔════╝
██████╔╝█████╗  ███████╗   ██║   ██║     █████╗  ███████╗███████╗
██╔══██╗██╔══╝  ╚════██║   ██║   ██║     ██╔══╝  ╚════██║╚════██║
██║  ██║███████╗███████║   ██║   ███████╗███████╗███████║███████║
╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   ╚══════╝╚══════╝╚══════╝╚══════╝
```

Adaptive API client with interactive mode, probing, simulation and export helpers.

Version: **v0.2.2-2-gb749d65**

---

## Install

### Debian / Ubuntu (APT)

```bash
echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless
```

### Homebrew (tap)

```bash
brew tap bspippi1337/homebrew-restless
brew install restless
```

---

## Demonstrations

### Interactive Mode (no args)

```bash
$ restless
```

```
🌀 Restless Interactive Mode
> probe https://api.example.com
> simulate https://api.example.com
> quit
```

### Probe

```bash
$ restless probe https://api.example.com
```

```json
{
  "url": "https://api.example.com",
  "methods": ["GET, POST"],
  "content_types": ["application/json"],
  "discovered_at": "2026-02-23T18:22:00Z"
}
```

### Simulate

```bash
$ restless simulate https://api.example.com
```

```
Method [GET]: POST
URL [https://api.example.com]:
Body: {"name":"pippi"}
```

### Smart Mode (guided)

```bash
$ restless smart https://api.example.com
```

```
[ Probing... ] ██████████ 100%
[ Profiling... ] ████████░░ 80%
[ Suggesting tools... ] ██████████ 100%
Ready.
```

### Export (examples)

```bash
$ restless export --format=har --out=req.har get https://api.example.com
$ restless export --format=curl --out=req.sh  post https://api.example.com
```

---

## Architecture

```
smartcmd
  │
  ├─ discover  → engine (suggest tools)
  ├─ simulator → guided request builder
  ├─ export    → json/md/har/curl outputs
  └─ app       → existing core (API, fuzzing, etc.)
```

---

## Distribution

- **APT**: published automatically to GitHub Pages via Actions (flat repo)
- **Brew**: formula auto-updated in tap repo from GitHub Release assets

Doctor:

```bash
./scripts/dist-doctor.sh
```

Release asset URL pattern:

```
https://github.com/bspippi1337/restless/releases/download/vv0.2.2-2-gb749d65/restless_v0.2.2-2-gb749d65_linux_amd64.tar.gz
```

