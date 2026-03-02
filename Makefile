.PHONY: tidy test run build build-all clean doctor lint

tidy:
	go mod tidy

test:
	go test $(shell go list ./... | grep -v /experiments/)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./... ; \
	else \
		echo "golangci-lint not installed, skipping lint." ; \
	fi

run:
	go run ./cmd/restless

build:
	mkdir -p bin
	go build -o bin/restless ./cmd/restless

build-all:
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build -o dist/restless_linux_amd64 ./cmd/restless
	GOOS=darwin  GOARCH=amd64 go build -o dist/restless_darwin_amd64 ./cmd/restless
	GOOS=windows GOARCH=amd64 go build -o dist/restless_windows_amd64.exe ./cmd/restless

clean:
	rm -rf bin dist build logs

doctor:
	go run ./cmd/restless doctor
