BINARY := soloenv
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/fleames/soloenv-cli/cmd.version=$(VERSION)

.PHONY: all build install test race vet fmt fmt-check check clean tidy run

all: check build

## build: compile the binary into ./$(BINARY)
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## install: install the binary into GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## test: run the test suite
test:
	go test -count=1 ./...

## race: run tests with the race detector
race:
	go test -race -count=1 ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format the code
fmt:
	gofmt -w .

## fmt-check: fail if code is not gofmt-ed
fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "not gofmt-ed:"; echo "$$out"; exit 1; fi

## check: fmt-check + vet + tests
check: fmt-check vet test

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf dist

## run: build and run (ARGS="up --protect")
run: build
	./$(BINARY) $(ARGS)

## help: list targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
