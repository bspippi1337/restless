#!/usr/bin/env bash
set -e
restless probe https://api.github.com
restless simulate https://api.github.com
restless smart https://api.github.com
