---
id: install
title: Installing Twirp
sidebar_label: Installation
---

Prerequisites

 * [Go](https://golang.org/): any one of the three latest major [releases of Go](https://golang.org/doc/devel/release.html). For installation instructions, see Go’s [Getting Started](https://golang.org/doc/install) guide.
 * [Protocol buffer](https://developers.google.com/protocol-buffers) compiler, `protoc` [version 3](https://developers.google.com/protocol-buffers/docs/proto3). For installation instructions, see [Protocol Buffer Compiler Installation](https://grpc.io/docs/protoc-installation/).

Install `protoc-gen-go` plugin for the protocol compiler (to generate `.pb.go` files):

```sh
go get google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

Install `protoc-gen-twirp` plugin for the protocol compiler (to generate `.twirp.go` files):

```sh
go get github.com/twitchtv/twirp/protoc-gen-twirp@latest
```

Go get installs the plugins in `$GOBIN` (defaults to `$GOPATH/bin`). They must be in your `$PATH` for the protocol compiler `protoc` to find them. You might need to add it to your path:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Older versions v5 and v7

Twirp v8 and above depend on the Protobuf APIV2. Older versions depend on Protobuf APIV1. See [Version Compatibility](version_matrix.md).

```sh
# Protobuf APIV1
go get github.com/golang/protobuf/protoc-gen-go@latest

# Twirp v5 or v7
go get github.com/twitchtv/twirp/protoc-gen-twirp@v7.2.0
```
