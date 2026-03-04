#!/usr/bin/env bash
set -e

echo
echo "=== restless API topology test ==="
echo

go run ./cmd/restless topology tools/mock_endpoints.txt

echo
echo "=== expected structure ==="
echo

cat <<TREE
API
└─ v1
  ├─ admin
  │ ├─ audit [GET]
  │ └─ stats [GET]
  ├─ health [GET]
  ├─ login [POST]
  ├─ logout [POST]
  └─ users [GET,POST]
    └─ {id} [GET]
TREE
