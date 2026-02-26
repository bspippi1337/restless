#!/usr/bin/env bash
set -e
go build -o /tmp/restless ./cmd/restless
/tmp/restless completion bash > docs/completion/restless.bash
/tmp/restless completion zsh > docs/completion/_restless
/tmp/restless completion fish > docs/completion/restless.fish
echo "Completion generated."
