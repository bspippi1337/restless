
# --- Reproducible build flags ---
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X main.version=$(VERSION)
BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

build:
	go build $(BUILD_FLAGS) -o restless ./cmd/restless

dist:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/restless_linux_amd64 ./cmd/restless
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/restless_darwin_amd64 ./cmd/restless
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/restless_windows_amd64.exe ./cmd/restless

checksum:
	cd dist && sha256sum * > checksums.txt
