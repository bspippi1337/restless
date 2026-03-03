APP=restless
VERSION=$(shell git describe --tags --always --dirty)

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(APP) ./cmd/restless

run:
	go run ./cmd/restless

fmt:
	go fmt ./...

tidy:
	go mod tidy

release-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(APP)-linux-amd64 ./cmd/restless

release-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(APP)-darwin-amd64 ./cmd/restless

release:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(APP)-linux-amd64 ./cmd/restless
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(APP)-darwin-amd64 ./cmd/restless
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(APP)-windows-amd64.exe ./cmd/restless
