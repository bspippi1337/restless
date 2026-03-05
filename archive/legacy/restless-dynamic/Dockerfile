FROM golang:1.22-alpine AS build
WORKDIR /app
COPY main.go .
RUN go build -o restless

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl
WORKDIR /app
COPY --from=build /app/restless .
EXPOSE 8080
CMD ["./restless"]
