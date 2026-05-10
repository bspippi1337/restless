FROM golang:1.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on go build \
    -trimpath \
    -buildvcs=false \
    -o /restless \
    ./cmd/restless

FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /restless /usr/bin/restless

RUN chmod +x /usr/bin/restless

ENTRYPOINT ["/usr/bin/restless"]
CMD ["--help"]
