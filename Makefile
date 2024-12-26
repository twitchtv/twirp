PATH := ${PWD}/bin:${PATH}
export GO111MODULE=off

all: setup test_all

.PHONY: setup generate test_all test test_clientcompat

setup:
	./check_protoc_version.sh
	# Install dependent tools via modules
	GO111MODULE=on GOBIN="$$PWD/bin" go install -v google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.1
	GO111MODULE=on GOBIN="$$PWD/bin" go install -v github.com/kisielk/errcheck@v1.8.0

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
