# Restless Manual

Version: v0.2.2-2-gb749d65

## CLI Help

```

```

## Smart Commands

- probe <url>        : inspect headers/method hints
- smart <url>        : profile + guided flow (expands over time)
- simulate <url>     : interactive request builder
- export ...         : export helpers (formats vary by build)

## Interactive Mode

Run without args:

```
restless
```

## Install via APT

```
echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update && sudo apt install restless
```

## Install via Homebrew

```
brew tap bspippi1337/homebrew-restless
brew install restless
```
