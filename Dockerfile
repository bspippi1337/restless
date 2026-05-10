FROM golang:1.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on go build \
    -trimpath \
    -buildvcs=false \
    -o /usr/local/bin/restless \
    ./cmd/restless

FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/restless /usr/local/bin/restless

RUN chmod 755 /usr/local/bin/restless && \
    /usr/local/bin/restless --version

ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["/usr/local/bin/restless"]
CMD ["--help"]
