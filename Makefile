PATH := ${PWD}/bin:${PATH}

all: setup test_all

.PHONY: setup generate test_all test test_clientcompat

setup:
	./check_protoc_version.sh
	GOBIN="$$PWD/bin" go install github.com/kisielk/errcheck@v1.20.0
	GOBIN="$$PWD/bin" go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0

generate:
	# Recompile and install generator
	GOBIN="$$PWD/bin" go install -v ./protoc-gen-twirp
	# Generate code from go:generate comments
	go generate ./...

test_all: setup test test_clientcompat

test: generate
	./bin/errcheck ./internal/twirptest
	go test -race ./...

test_clientcompat: generate
	GOBIN="$$PWD/bin" go install ./clientcompat
	GOBIN="$$PWD/bin" go install ./clientcompat/gocompat
	./bin/clientcompat -client ./bin/gocompat
