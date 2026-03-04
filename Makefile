.PHONY: tidy build

tidy:
	go mod tidy

#test:
#	go test ./...

build:
	go build -o build/restless ./cmd/restless

#lint:
#	golangci-lint run ./...

clean:
	rm -rf build
