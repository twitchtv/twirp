PATH := ${PWD}/_tools/bin:${PWD}/bin:${PWD}/ENV/bin:${PATH}
export GO111MODULE=off

all: setup test_all

.PHONY: setup generate test_all test test_clients test_go_client test_python_client

setup:
	./check_protoc_version.sh
	GOPATH=$(CURDIR)/_tools GOBIN=$(CURDIR)/_tools/bin go get github.com/twitchtv/retool
	./_tools/bin/retool build

generate:
	# Recompile and install generator
	GOBIN="$$PWD/bin" go install -v ./protoc-gen-twirp
	GOBIN="$$PWD/bin" go install -v ./protoc-gen-twirp_python
	# Generate code from go:generate comments
	go generate ./...

test_all: setup test test_clients

test: generate
	./_tools/bin/errcheck ./internal/twirptest
	go test -race $(shell GO111MODULE=off go list ./... | grep -v /vendor/ | grep -v /_tools/)

test_clients: test_go_client test_python_client

test_go_client: generate build/clientcompat build/gocompat
	./build/clientcompat -client ./build/gocompat

test_python_client: generate build/clientcompat build/pycompat
	./build/clientcompat -client ./build/pycompat


# For clientcompat and testing Python
./build:
	mkdir build

./build/gocompat: ./build
	go build -o build/gocompat ./clientcompat/gocompat

./build/clientcompat: ./build
	go build -o build/clientcompat ./clientcompat

./build/venv: ./build
	virtualenv ./build/venv

./build/venv/bin/pycompat.py: ./build/venv
	./build/venv/bin/pip install --upgrade ./clientcompat/pycompat

./build/pycompat: ./build/venv/bin/pycompat.py
	cp ./clientcompat/pycompat/pycompat.sh ./build/pycompat
	chmod +x ./build/pycompat
