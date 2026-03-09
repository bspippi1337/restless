#!/usr/bin/env bash
set -e

echo "== Restless Repo Doctor =="

echo "checking go modules..."
go mod tidy

echo "checking formatting..."
gofmt -l .

echo "running vet..."
go vet ./...

echo "building..."
go build ./cmd/restless

echo "repo healthy"
