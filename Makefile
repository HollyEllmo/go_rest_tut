.PHONY: build test run clean

build:
	@mkdir -p bin
	@go build -o bin/go_rest_tut cmd/main.go

test:
	@go test -v ./...

run: build
	@./bin/go_rest_tut

clean:
	@rm -rf bin/