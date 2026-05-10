FROM golang:1.24 AS builder

WORKDIR /src
COPY . .

RUN set -eux; \
    TARGET="."; \
    [ -f ./cmd/restless/main.go ] && TARGET="./cmd/restless"; \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /tmp/restless \
      "$TARGET"; \
    chmod +x /tmp/restless

FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /tmp/restless /usr/local/bin/restless

RUN chmod +x /usr/local/bin/restless

ENTRYPOINT ["/usr/local/bin/restless"]
CMD ["--help"]
