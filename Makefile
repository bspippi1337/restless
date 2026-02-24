BINARY := restless
VERSION ?= 4.0.4
PKG := github.com/bspippi1337/restless/internal/version

.PHONY: build test clean build-all tidy fmt doctor

build:
	CGO_ENABLED=0 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o $(BINARY) ./cmd/restless

build-all:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_linux_amd64   ./cmd/restless
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_linux_arm64   ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_darwin_amd64  ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_darwin_arm64  ./cmd/restless
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_windows_amd64.exe ./cmd/restless

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)
	rm -rf dist

doctor: tidy fmt test build
	@echo "Doctor OK"
