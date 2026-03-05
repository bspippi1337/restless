FROM golang:1.22-alpine AS build
WORKDIR /app
COPY main.go .
RUN go mod init restless && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o restless

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/restless .
EXPOSE 8080
CMD ["./restless"]
