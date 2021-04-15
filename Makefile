RETOOL=$(CURDIR)/_tools/bin/retool
PATH := ${PWD}/bin:${PWD}/ENV/bin:${PATH}
DOCKER_RELEASE_IMAGE := golang:1.12.0-stretch
.DEFAULT_GOAL := all

all: setup test_all

.PHONY: test test_all test_core test_clients test_go_client test_python_client generate release_gen

generate:
	GO111MODULE=off GOBIN="$$PWD/bin" go install -v ./protoc-gen-twirp
	GO111MODULE=off GOBIN="$$PWD/bin" go install -v ./protoc-gen-twirp_python
	GO111MODULE=off PATH="$$PWD/bin:$$PWD/_tools/bin:$$PATH" go generate ./...

gen:
	GO111MODULE=off PATH="$$PWD/bin:$$PWD/_tools/bin:$$PATH" go generate ./...

test_all: setup test_core test_clients

test_core: generate
	GO111MODULE=off PATH="$$PWD/bin:$$PWD/_tools/bin:$$PATH" errcheck -blank ./internal/twirptest
	GO111MODULE=off go test -race $(shell GO111MODULE=off go list ./... | grep -v /vendor/ | grep -v /_tools/)

test_clients: test_go_client test_python_client

test_go_client: generate build/clientcompat build/gocompat
	./build/clientcompat -client ./build/gocompat

test_python_client: generate build/clientcompat build/pycompat
	./build/clientcompat -client ./build/pycompat

setup:
	./install_proto.bash
	GO111MODULE=off GOPATH=$(CURDIR)/_tools GOBIN=$(CURDIR)/_tools/bin go get github.com/twitchtv/retool
	$(RETOOL) build

release_gen:
	git clean -xdf
	docker run \
		--volume "$(CURDIR):/go/src/github.com/twitchtv/twirp" \
		--workdir "/go/src/github.com/twitchtv/twirp" \
		$(DOCKER_RELEASE_IMAGE) \
		internal/release_gen.sh

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
