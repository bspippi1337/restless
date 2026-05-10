FROM golang:1.24 AS builder

WORKDIR /src

COPY . .

RUN mkdir -p /out

RUN if [ -f ./cmd/restless/main.go ]; then \
        echo "[*] building ./cmd/restless"; \
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -trimpath -ldflags="-s -w" \
        -o /out/restless ./cmd/restless ; \
    else \
        echo "[*] building repo root"; \
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -trimpath -ldflags="-s -w" \
        -o /out/restless . ; \
    fi

RUN test -f /out/restless
RUN chmod +x /out/restless
RUN ls -lah /out/restless

FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/restless /usr/local/bin/restless

RUN chmod +x /usr/local/bin/restless && \
    test -x /usr/local/bin/restless && \
    ls -lah /usr/local/bin/restless

ENV PATH="/usr/local/bin:/usr/bin:/bin"

ENTRYPOINT ["/usr/local/bin/restless"]
CMD ["--help"]
