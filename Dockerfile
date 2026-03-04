# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/restless ./cmd/restless

FROM alpine:latest
RUN adduser -D -H -s /sbin/nologin restless
COPY --from=build /out/restless /usr/local/bin/restless
USER restless
ENTRYPOINT ["restless"]
