DEP_VERSION := 0.5.0
PROTOC_VERSION := 3.7.0
ERRCHECK_VERSION := v1.2.0
PROTOC_GEN_GO_VERSION := v1.3.0
PROTOC_GEN_GOFAST_VERSION := v1.2.1

SHELL := /bin/bash -o pipefail
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_BIN = $(TMP)/bin
TMP_VENV = $(TMP)/venv

$(TMP_BIN):
	@mkdir -p "$(TMP_BIN)"

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
$(DEP): $(TMP_BIN)
	curl -sSL "https://github.com/golang/dep/releases/download/v$(DEP_VERSION)/dep-$(DEP_OS)-$(DEP_ARCH)" -o "$(DEP)"
	@chmod +x "$(DEP)"

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
$(PROTOC): $(TMP_BIN)
	curl -sSL "https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip" -o "$(TMP)/protoc.zip"
	@rm -rf "$(TMP)/include/google/protobuf"
	cd "$(TMP)"; unzip -o protoc.zip
	@rm -f "$(TMP)/readme.txt"

ERRCHECK := $(TMP_BIN)/errcheck
$(ERRCHECK):
	$(eval ERRCHECK_TMP := $(shell mktemp -d))
	@cd $(ERRCHECK_TMP); echo module tmp > go.mod; GO111MODULE=on go get github.com/kisielk/errcheck@$(ERRCHECK_VERSION)
	@rm -rf $(ERRCHECK_TMP)

PROTOC_GEN_GO := $(TMP_BIN)/protoc-gen-go
$(PROTOC_GEN_GO):
	$(eval PROTOC_GEN_GO_TMP := $(shell mktemp -d))
	@cd $(PROTOC_GEN_GO_TMP); echo module tmp > go.mod; GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@rm -rf $(PROTOC_GEN_GO_TMP)

PROTOC_GEN_GOFAST := $(TMP_BIN)/protoc-gen-gofast
$(PROTOC_GEN_GOFAST):
	$(eval PROTOC_GEN_GOFAST_TMP := $(shell mktemp -d))
	@cd $(PROTOC_GEN_GOFAST_TMP); echo module tmp > go.mod; GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gofast@$(PROTOC_GEN_GOFAST_VERSION)
	@rm -rf $(PROTOC_GEN_GOFAST_TMP)

export GOBIN := $(abspath $(TMP_BIN))
export PATH := $(GOBIN):$(PATH)

$(TMP_BIN)/gocompat: $(TMP_BIN)
	go build -o $(TMP_BIN)/gocompat ./clientcompat/gocompat

$(TMP_BIN)/clientcompat: $(TMP_BIN)
	go build -o $(TMP_BIN)/clientcompat ./clientcompat

$(TMP_VENV):
	@mkdir -p $(shell dirname $(TMP_VENV))
	virtualenv $(TMP_VENV)

$(TMP_VENV)/bin/pycompat.py: $(TMP_VENV)
	$(TMP_VENV)/bin/pip install --upgrade ./clientcompat/pycompat

$(TMP)/pycompat: $(TMP_VENV)/bin/pycompat.py
	cp ./clientcompat/pycompat/pycompat.sh $(TMP)/pycompat
	@chmod +x $(TMP)/pycompat

.DEFAULT_GOAL := all

.PHONY: all
all: test_all

.PHONY: generate
generate: $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GOFAST)
	go install ./protoc-gen-twirp ./protoc-gen-twirp_python
	go generate ./...

.PHONY: vendor_update
vendor_update: $(DEP)
	dep ensure -update -v

.PHONY: clean
clean:
	rm -rf $(TMP_BASE)

.PHONY: test_all
test_all: test_core test_clients

.PHONY: test_core
test_core: generate $(ERRCHECK)
	errcheck -blank ./internal/twirptest
	go test -race $(shell go list ./... | grep -v /vendor/)

.PHONY: test_clients
test_clients: test_go_client test_python_client

.PHONY: test_go_client
test_go_client: generate $(TMP_BIN)/clientcompat $(TMP_BIN)/gocompat
	./$(TMP_BIN)/clientcompat -client ./$(TMP_BIN)/gocompat

.PHONY: test_python_client
test_python_client: generate $(TMP_BIN)/clientcompat $(TMP)/pycompat
	./$(TMP_BIN)/clientcompat -client ./$(TMP)/pycompat
