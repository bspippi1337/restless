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
      "$TARGET"

FROM debian:stable-slim

ENTRYPOINT ["/usr/local/bin/restless"]
CMD ["--help"]

COPY --from=builder /tmp/restless /usr/local/bin/restless

RUN chmod +x /usr/local/bin/restless
