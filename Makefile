SHELL := /bin/bash -o pipefail
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_BIN = $(TMP)/bin

DEP_VERSION := 0.5.0
DEP := $(TMP_BIN)/dep
ifeq ($(UNAME_OS),Darwin)
DEP_OS := darwin
endif
ifeq ($(UNAME_OS),Linux)
DEP_OS = linux
endif
ifeq ($(UNAME_ARCH),x86_64)
DEP_ARCH := amd64
endif
$(DEP):
	@mkdir -p "$(TMP_BIN)"
	curl -sSL "https://github.com/golang/dep/releases/download/v$(DEP_VERSION)/dep-$(DEP_OS)-$(DEP_ARCH)" -o "$(DEP)"
	@chmod +x "$(DEP)"

PROTOC_VERSION := 3.7.0
PROTOC := $(TMP_BIN)/protoc
ifeq ($(UNAME_OS),Darwin)
PROTOC_OS := osx
endif
ifeq ($(UNAME_OS),Linux)
PROTOC_OS = linux
endif
ifeq ($(UNAME_ARCH),x86_64)
PROTOC_ARCH := x86_64
endif
$(PROTOC):
	@mkdir -p "$(TMP_BIN)"
	curl -sSL "https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip" -o "$(TMP)/protoc.zip"
	cd "$(TMP)"; unzip protoc.zip
	@rm -f "$(TMP)/readme.txt"

ERRCHECK_VERSION := v1.2.0
ERRCHECK := $(TMP_BIN)/errcheck
$(ERRCHECK):
	$(eval ERRCHECK_TMP := $(shell mktemp -d))
	@cd $(ERRCHECK_TMP); echo module tmp > go.mod; GO111MODULE=on go get github.com/kisielk/errcheck@$(ERRCHECK_VERSION)
	@rm -rf $(ERRCHECK_TMP)

PROTOC_GEN_GO_VERSION := v1.3.0
PROTOC_GEN_GO := $(TMP_BIN)/protoc-gen-go
$(PROTOC_GEN_GO):
	$(eval PROTOC_GEN_GO_TMP := $(shell mktemp -d))
	@cd $(PROTOC_GEN_GO_TMP); echo module tmp > go.mod; GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@rm -rf $(PROTOC_GEN_GO_TMP)

PROTOC_GEN_GOFAST_VERSION := v1.2.1
PROTOC_GEN_GOFAST := $(TMP_BIN)/protoc-gen-gofast
$(PROTOC_GEN_GOFAST):
	$(eval PROTOC_GEN_GOFAST_TMP := $(shell mktemp -d))
	@cd $(PROTOC_GEN_GOFAST_TMP); echo module tmp > go.mod; GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gofast@$(PROTOC_GEN_GOFAST_VERSION)
	@rm -rf $(PROTOC_GEN_GOFAST_TMP)

export GOBIN := $(abspath $(TMP_BIN))
export PATH := $(GOBIN):$(PATH)

.DEFAULT_GOAL := all

all: test_all

.PHONY: test test_all test_core test_clients test_go_client test_python_client generate vendor_update clean

# Phony commands:

generate: $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GOFAST)
	go install ./protoc-gen-twirp ./protoc-gen-twirp_python
	go generate ./...

vendor_update: $(DEP)
	dep ensure -update -v

clean:
	rm -rf $(TMP_BASE)

test_all: test_core test_clients

test_core: generate $(ERRCHECK)
	errcheck -blank ./internal/twirptest
	go test -race $(shell go list ./... | grep -v /vendor/)

test_clients: test_go_client test_python_client

test_go_client: generate build/clientcompat build/gocompat
	./build/clientcompat -client ./build/gocompat

test_python_client: generate build/clientcompat build/pycompat
	./build/clientcompat -client ./build/pycompat

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
