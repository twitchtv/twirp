PATH := ${PWD}/_tools/bin:${PWD}/bin:${PWD}/ENV/bin:${PATH}
export GO111MODULE=off

all: setup test_all

.PHONY: setup generate test_all test test_clientcompat build_clientcompact

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
	go test -race ./...

build_clientcompact:
	mkdir -p clientcompat/bin
	go build -o clientcompat/bin/clientcompat ./clientcompat
	go build -o clientcompat/bin/gocompat ./clientcompat/gocompat

test_clientcompat: generate build_clientcompact
	./clientcompat/bin/clientcompat -client ./clientcompat/bin/gocompat
