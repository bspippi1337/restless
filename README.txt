
Restless merge patch (discover + map + graph + svg + completion)
=============================================================

Apply:
  unzip -o restless_merge_patch.zip -d .

IMPORTANT: remove legacy discovery engine if present (prevents Graph/Run redeclare):
  rm -f internal/core/discover/discover.go

Build:
  go mod tidy
  go build -o build/restless ./cmd/restless

Run:
  ./build/restless discover https://api.github.com
  ./build/restless map
  ./build/restless map --tree=false
  ./build/restless graph
  ./build/restless graph --svg
  ./build/restless completion
