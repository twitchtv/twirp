RETOOL=$(CURDIR)/_tools/bin/retool
PATH := ${PWD}/bin:${PWD}/ENV/bin:${PATH}
.DEFAULT_GOAL := all

all: setup test_all

.PHONY: test test_all test_core test_clients test_go_client test_python_client generate

# Phony commands:
generate:
	PATH=$(CURDIR)/_tools/bin:$(PATH) GOBIN="${PWD}/bin" go install -v ./protoc-gen-...
	$(RETOOL) do go generate ./...

test_all: setup test_core test_clients

test_core: generate
	$(RETOOL) do errcheck -blank ./internal/twirptest
	go test -race $(shell go list ./... | grep -v /vendor/ | grep -v /_tools/)

test_clients: test_go_client test_python_client

test_go_client: generate build/clientcompat build/gocompat
	./build/clientcompat -client ./build/gocompat

test_python_client: generate build/clientcompat build/pycompat
	./build/clientcompat -client ./build/pycompat

setup:
	./install_proto.bash
	GOPATH=$(CURDIR)/_tools go install github.com/twitchtv/retool/...
	$(RETOOL) build

# Actual files for testing clients:
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
