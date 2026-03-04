Restless Supersolid CLI Patch
=============================

Drop these files into your repo root and commit.

Steps:

1. Extract zip in repo root
2. Ensure old discovery engine is removed:
   rm internal/core/discover/discover.go  (if present)

3. Build:

   go mod tidy
   go build -o build/restless ./cmd/restless

4. Test:

   ./build/restless discover https://api.github.com
   ./build/restless graph
   ./build/restless graph --svg
   ./build/restless completion

Completion:

bash:
source dist/completion/restless.bash

zsh:
fpath+=(dist/completion)
autoload -Uz compinit && compinit