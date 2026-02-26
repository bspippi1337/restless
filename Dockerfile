# ---------- STAGE 1: builder ----------
FROM golang:1.22-bookworm AS builder

# install useful debugging tools
RUN apt-get update && apt-get install -y \
    git bash ca-certificates curl make file \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# copy entire repo into container
COPY . .

# ---------- AUTOPILOT SCRIPT ----------
# This script attempts to automatically understand the project
RUN cat <<'EOF' > /autopilot.sh
#!/usr/bin/env bash
set -euo pipefail

echo "🧠 Autopilot engaged"
echo "--------------------------------------------------"

echo "📍 Go version:"
go version

echo "📍 Project tree:"
ls -la

echo "📍 Searching for go.mod..."
if [ ! -f go.mod ]; then
  echo "❌ No go.mod found. Not a Go project."
  exit 1
fi

echo "📦 Running go mod tidy"
go mod tidy

echo "🧹 Running go fmt"
gofmt -w .

echo "🔎 Running go vet"
go vet ./... || true

echo "🧪 Running tests (if any)"
go test ./... || true

echo "🔍 Discovering main packages..."
MAINS=$(grep -rl "package main" --include="*.go" . | sed 's|/[^/]*$||' | sort -u)

if [ -z "$MAINS" ]; then
  echo "❌ No main packages found."
  exit 1
fi

echo "🚀 Found main packages:"
echo "$MAINS"

mkdir -p /out

echo "🔨 Building binaries..."
for dir in $MAINS; do
    name=$(basename "$dir")
    echo "   building $name from $dir"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -ldflags "-s -w" -o "/out/$name" "$dir" || true
done

echo "📦 Built binaries:"
ls -lah /out

echo "✅ Autopilot finished successfully"
EOF

RUN chmod +x /autopilot.sh

# run the autopilot build
RUN /autopilot.sh


# ---------- STAGE 2: runtime image ----------
FROM debian:bookworm-slim

WORKDIR /app

# copy all built binaries
COPY --from=builder /out /usr/local/bin

RUN chmod +x /usr/local/bin/* || true

CMD ["bash"]
