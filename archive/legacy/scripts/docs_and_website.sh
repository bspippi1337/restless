#!/usr/bin/env bash
set -euo pipefail

VERSION="4.0.4"

echo "==> Rebuilding documentation and website"

# -------------------------
# Docs Structure
# -------------------------

rm -rf docs
mkdir -p docs

cat > docs/OVERVIEW.md <<EOT
# Restless Overview

Restless is a terminal-first API workbench designed for engineers who prefer precision, automation, and composability over graphical tooling.

Version: $VERSION

## Core Capabilities

- OpenAPI import and execution
- Endpoint discovery
- OperationId support
- Interactive parameter prompting
- Profiles (environment base URLs)
- Session templating
- Curl generation
- Artifact export
- Built-in benchmarking
- Latency histogram
- Strict CI mode

Restless is not a GUI replacement.
It is an execution engine.
EOT

cat > docs/CLI.md <<'EOT'
# CLI Usage

## Raw Request

restless -X POST -url https://api.example.com -d '{"hello":"ground friend"}'

## OpenAPI

restless openapi import spec.json
restless openapi ls
restless openapi endpoints <id>
restless openapi run <id> GET /path

## Profiles

restless profile set dev base=https://api.example.com
restless profile use dev
restless profile ls

## Preferences

restless prefs show
restless prefs set color=on
EOT

cat > docs/ARCHITECTURE.md <<'EOT'
# Architecture

Restless is structured around modular execution layers:

cmd/restless        → CLI router
internal/core/app   → module registry
internal/modules/
    openapi         → spec engine
    session         → templating
    bench           → load testing
    export          → artifacts
internal/version    → version injection

Design principles:

- Deterministic execution
- CI-friendly
- Static builds
- Zero external runtime dependencies
EOT

cat > docs/RELEASE.md <<EOT
# Release $VERSION

This release formalizes Restless as a stable OpenAPI-first execution engine.

Key additions:

- Profile-based base injection
- Strict mode enforcement
- Interactive param prompting
- Latency histogram
- Version injection via ldflags
- Cross-platform static builds
EOT

# -------------------------
# Rewrite README
# -------------------------

cat > README.md <<EOT
# Restless ⚡

Terminal-First API Workbench

Version: $VERSION

Restless is a modular OpenAPI-aware execution engine built for shell-native development.

## Install

go build -o restless ./cmd/restless

## Example

restless openapi import petstore.json  
restless openapi run <id> GET /pets  

## Philosophy

- Scriptable
- Deterministic
- CI-native
- Modular

See /docs for full documentation.
EOT

# -------------------------
# Website
# -------------------------

rm -rf site
mkdir -p site

cat > site/index.html <<EOT
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Restless</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
<header>
<h1>Restless</h1>
<p>Terminal-First API Workbench</p>
<span class="version">v$VERSION</span>
</header>

<section>
<h2>What is Restless?</h2>
<p>
Restless is a modular, OpenAPI-first execution engine built for developers who prefer precision over pixels.
</p>
</section>

<section>
<h2>Core Features</h2>
<ul>
<li>OpenAPI import & execution</li>
<li>Profiles & environment isolation</li>
<li>Session templating</li>
<li>Strict CI mode</li>
<li>Latency histogram benchmarking</li>
<li>Cross-platform static binaries</li>
</ul>
</section>

<section>
<h2>Install</h2>
<pre>go build -o restless ./cmd/restless</pre>
</section>

<footer>
<p>Version $VERSION • MIT License</p>
</footer>

</body>
</html>
EOT

cat > site/style.css <<'EOT'
body {
    font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif;
    background: #0f1115;
    color: #e6edf3;
    margin: 0;
    padding: 0;
}

header {
    text-align: center;
    padding: 4rem 2rem;
    background: #151922;
}

h1 {
    font-size: 3rem;
    margin: 0;
}

.version {
    display: inline-block;
    margin-top: 1rem;
    padding: 0.3rem 0.8rem;
    background: #1f6feb;
    border-radius: 6px;
    font-size: 0.9rem;
}

section {
    padding: 3rem 2rem;
    max-width: 800px;
    margin: auto;
}

pre {
    background: #161b22;
    padding: 1rem;
    border-radius: 6px;
    overflow-x: auto;
}

footer {
    text-align: center;
    padding: 2rem;
    background: #151922;
    font-size: 0.8rem;
}
EOT

echo "==> Documentation rebuilt"
echo "==> Website generated in /site"
echo "Ready for GitHub Pages"
