#!/usr/bin/env bash

set -euo pipefail

# protoc_gen.sh foo.proto will compile foo.proto. Should be run in the same
# directory as its input. Handles multi-element GOPATHs so it works with retool.

# Append '/src' to every element in GOPATH.
PROTOPATH=${GOPATH/://src:}/src

protoc --twirp_out=paths=source_relative:. --go_out=paths=source_relative:. "$@"
protoc --proto_path="${PROTOPATH}:." --python_out=. --twirp_python_out=. "$@"
