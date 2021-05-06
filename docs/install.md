---
id: install
title: Installing Twirp
sidebar_label: Installation
---

Prerequisites

 * [Go](https://golang.org/): any one of the three latest major [releases of Go](https://golang.org/doc/devel/release.html). For installation instructions, see Go’s [Getting Started](https://golang.org/doc/install) guide.
 * [Protocol buffer](https://developers.google.com/protocol-buffers) compiler, `protoc` [version 3](https://developers.google.com/protocol-buffers/docs/proto3). For installation instructions, see [Protocol Buffer Compiler Installation](https://grpc.io/docs/protoc-installation/).

Install `protoc-gen-go` plugin for the protocol compiler (to generate `.pb.go` files):

```
go get google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

NOTE: Older versions of Twirp v5 and v7 use the older Go plugin: `go get github.com/golang/protobuf/protoc-gen-go`

Install `protoc-gen-twirp` plugin for the protocol compiler (to generate `.twirp.go` files):

```
go get github.com/twitchtv/twirp/protoc-gen-twirp@latest
```

Go get installs the plugins in `$GOBIN` (defaults to `$GOPATH/bin`). They must be in your `$PATH` for the protocol compiler `protoc` to find them. You might need to add it to your path:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

