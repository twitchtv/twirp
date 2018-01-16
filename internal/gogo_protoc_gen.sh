#!/usr/bin/env bash

set -euo pipefail

# gogo_protoc_gen.sh foo.proto will compile foo.proto using
# github.com/gogo/protobuf/protoc-gen-gofast, an alternative generator used
# sometimes at Twitch.. Should be run in the same directory as its input.
# Handles multi-element GOPATHs so it works with retool.

# Append '/src' to every element in GOPATH.
PROTOPATH=${GOPATH/://src:}/src

protoc --proto_path="${PROTOPATH}:." --twirp_out=. --gofast_out=. "$@"
