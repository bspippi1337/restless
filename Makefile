BINARY := restless

.PHONY: build test clean

build:
	CGO_ENABLED=0 go build -o $(BINARY) ./cmd/restless

test:
	go test ./...

clean:
	rm -f $(BINARY)

build-all:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o dist/restless_linux_amd64   ./cmd/restless
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -o dist/restless_linux_arm64   ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o dist/restless_darwin_amd64  ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o dist/restless_darwin_arm64  ./cmd/restless
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/restless_windows_amd64.exe ./cmd/restless
