.PHONY: tidy test run build build-all clean doctor

tidy:
	go mod tidy

test:
	go test ./...

run:
	go run ./cmd/restless

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/restless ./cmd/restless

build-all:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/restless_linux_amd64 ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/restless_darwin_amd64 ./cmd/restless
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/restless_windows_amd64.exe ./cmd/restless

clean:
	rm -rf bin dist build logs

doctor:
	go run ./cmd/restless doctor
