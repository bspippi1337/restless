FROM golang:1.24-alpine AS build

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/restless \
    ./cmd/restless


FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/restless .

ENTRYPOINT ["./restless"]
