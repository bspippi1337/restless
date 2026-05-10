FROM golang:1.24 AS builder

WORKDIR /src
COPY . .

RUN set -eux; \
    mkdir -p /out; \
    TARGET="."; \
    [ -f ./cmd/restless/main.go ] && TARGET="./cmd/restless"; \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/restless \
      "$TARGET"; \
    chmod +x /out/restless; \
    ls -lah /out/restless

FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/restless /usr/local/bin/restless

RUN set -eux; \
    ls -lah /usr/local/bin/restless; \
    chmod +x /usr/local/bin/restless; \
    /usr/local/bin/restless --help >/dev/null 2>&1 || true

ENV PATH="/usr/local/bin:/usr/bin:/bin"

ENTRYPOINT ["/usr/local/bin/restless"]
CMD ["--help"]
