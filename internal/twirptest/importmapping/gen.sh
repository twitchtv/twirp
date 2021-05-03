#!/usr/bin/env bash
set -euo pipefail

protoc --go_out=. --twirp_out=. \
    --go_opt=paths=source_relative \
    --go_opt=My/y.proto=github.com/twitchtv/twirp/internal/twirptest/importmapping/y \
    --twirp_opt=paths=source_relative \
    --twirp_opt=My/y.proto=github.com/twitchtv/twirp/internal/twirptest/importmapping/y \
    y/y.proto

protoc --go_out=. --twirp_out=. \
    --go_opt=paths=source_relative \
    --go_opt=My/y.proto=github.com/twitchtv/twirp/internal/twirptest/importmapping/y \
    --go_opt=Mx/x.proto=github.com/twitchtv/twirp/internal/twirptest/importmapping/x \
    --twirp_opt=paths=source_relative \
    --twirp_opt=My/y.proto=github.com/twitchtv/twirp/internal/twirptest/importmapping/y \
    --twirp_opt=Mx/x.proto=github.com/twitchtv/twirp/internal/twirptest/importmapping/x \
    x/x.proto
