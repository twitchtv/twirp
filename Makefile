PATH := ${PWD}/_tools/bin:${PWD}/bin:${PWD}/ENV/bin:${PATH}
export GO111MODULE=off

all: setup test_all

.PHONY: setup generate test_all test test_clientcompat

setup:
	./check_protoc_version.sh
	GOPATH=$(CURDIR)/_tools GOBIN=$(CURDIR)/_tools/bin go get github.com/twitchtv/retool
	./_tools/bin/retool build

generate:
	# Recompile and install generator
	GOBIN="$$PWD/bin" go install -v ./protoc-gen-twirp
	# Generate code from go:generate comments
	go generate ./...

test_all: setup test test_clientcompat

test: generate
	./_tools/bin/errcheck ./internal/twirptest
	go test -race $(shell GO111MODULE=off go list ./... | grep -v /vendor/ | grep -v /_tools/)

test_clientcompat: generate build/clientcompat build/gocompat
	./build/clientcompat -client ./build/gocompat

./build:
	mkdir build

./build/gocompat: ./build
	go build -o build/gocompat ./clientcompat/gocompat

./build/clientcompat: ./build
	go build -o build/clientcompat ./clientcompat
