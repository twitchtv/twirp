# Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may not
# use this file except in compliance with the License. A copy of the License is
# located at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# or in the "license" file accompanying this file. This file is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

M = $(shell printf "\033[34;1m▶\033[0m")

export GOBIN = $(CURDIR)/_bin
export PROTOPATH = $(GOPATH)/src
export BINDIR := $(GOBIN)
export PATH := $(GOBIN):$(PATH)

protoc_gen_go := $(GOBIN)/protoc-gen-go
protoc_gen_go_src := vendor/github.com/golang/protobuf/protoc-gen-go

protoc_gen_twirp := $(GOBIN)/protoc-gen-twirp
protoc_gen_twirp_src := vendor/github.com/twitchtv/twirp/protoc-gen-twirp

build: $(info $(M) Building project...)
	@ CGO_ENABLED=0 go build -o /go/bin/example ./cmd/example/*.go
.PHONY: build

gen-twirp: $(protoc_gen_go) $(protoc_gen_twirp)
	$(info $(M) Generating twirp files...)
	@protoc --proto_path=$(GOBIN):. --twirp_out=. --go_out=. ./rpc/example/service.proto
.PHONY: gen-twirp

$(protoc_gen_go): $(protoc_gen_go_src)
	@go install ./$^

$(protoc_gen_twirp): $(protoc_gen_twirp_src)
	@go install ./$^

local:
	$(info $(M) Starting Development server...)
	go run ./cmd/example/main.go

docker-image:
	$(info $(M) Building app image...)
	docker build -t example .

docker-container: docker-image
	$(info $(M) Running docker application container...)
	docker run -p 8080:8080 example:latest
