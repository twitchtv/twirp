PATH := ${PWD}/bin:${PWD}/ENV/bin:${PATH}
DOCKER_RELEASE_IMAGE := golang:1.14.0-stretch
.DEFAULT_GOAL := all

TOOLS_BIN ?= $(CURDIR)/_tools/bin
PROTOC_PATH ?= $(CURDIR)/_tools

PROTOBUF_VERSION ?= 3.11.0

ifeq ($(UNAME_S),Darwin)
	PROTOC_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOBUF_VERSION)/protoc-$(PROTOBUF_VERSION)-osx-x86_64.zip
else
	PROTOC_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOBUF_VERSION)/protoc-$(PROTOBUF_VERSION)-linux-x86_64.zip
endif

all: protoc setup test_all

.PHONY: protoc test test_all test_core test_clients test_go_client test_python_client generate release_gen

protoc:
	@if [ ! -d $(PROTOC_PATH) ]; then\
		mkdir -p $(PROTOC_PATH)/bin;\
	fi

	@if [ ! -f $(PROTOC_PATH)/bin/protoc ]; then\
		echo "Installing $(PROTOC_URL) to $(PROTOC_PATH)";\
		curl -o protoc.zip -sSL $(PROTOC_URL);\
		unzip -u protoc.zip -d $(PROTOC_PATH);\
		rm -rf protoc.zip;\
	fi

# Phony commands:
generate:
	GOBIN="${PWD}/bin" go install -v ./protoc-gen-...
	PATH=$(TOOLS_BIN):$(PATH) go generate ./...

test_all: setup test_core test_clients

test_core: generate
	GOBIN=$(TOOLS_BIN) errcheck -blank ./internal/twirptest
	go test -race $(shell go list ./... | grep -v /vendor/ | grep -v /_tools/)

test_clients: test_go_client test_python_client

test_go_client: generate build/clientcompat build/gocompat
	./build/clientcompat -client ./build/gocompat

test_python_client: generate build/clientcompat build/pycompat
	./build/clientcompat -client ./build/pycompat

setup:
	GOBIN=$(TOOLS_BIN) go install github.com/golang/protobuf/protoc-gen-go
	GOBIN=$(TOOLS_BIN) go install github.com/kisielk/errcheck
	GOBIN=$(TOOLS_BIN) go install github.com/gogo/protobuf/protoc-gen-gofast

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
