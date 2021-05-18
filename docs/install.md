---
id: install
title: Installing Twirp
sidebar_label: Installation
---

## Runtime Library

The runtime library package `github.com/twitchtv/twirp` contains common types like `twirp.Error`. If you are only importing Twirp clients from other services, you only need to import the twirp package and the protobuf APIv2 dependency (`google.golang.org/protobuf`).

If the Twirp client was generated with older versions of Twirp (v5, v7), then you need to import the older protobuf APIv1 dependency (`github.com/golang/protobuf`).


## Code Generator

You need to install `go` and the `protoc` compiler in your system. Then, install the protoc plugins `protoc-gen-twirp` and `protoc-gen-go` to generate Go code.


### Prerequisites

 * [Go](https://golang.org/): Twirp works well with any one of the three latest major [releases of Go](https://golang.org/doc/devel/release.html). For installation instructions, see Go’s [Getting Started](https://golang.org/doc/install) guide.
 * [Protocol buffer](https://developers.google.com/protocol-buffers) compiler, `protoc` [version 3](https://developers.google.com/protocol-buffers/docs/proto3). For installation instructions, see [Protocol Buffer Compiler Installation](https://grpc.io/docs/protoc-installation/) (For example in MacOS: `brew install protobuf`).


### Define tools.go for versioning in go.mod

You should track the Twirp and Protobuf versions like any other go-based tool (e.g. `stringer`). The currently recommended approach is to track the tool's version in your module's `go.mod` file (See ["Go Modules by Example" walkthrough](https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md)). For example, a `tools.go` file may look like this:

```go
// +build tools

package tools

import (
        _ "google.golang.org/protobuf/cmd/protoc-gen-go"
        _ "github.com/twitchtv/twirp/protoc-gen-twirp"
)
```

### Install Twirp and Protobuf Generators

Set `GOBIN` (see [go help environment](https://golang.org/cmd/go/#hdr-Environment_variables)) to define where the tool dependencies will be installed. For example, if you have a `/bin` folder in your project:

```sh
export GOBIN=$PWD/bin
```

The installed packages need to be accessible by the `protoc` compiler. You might need to add GOBIN to your PATH:

```sh
export PATH=$GOBIN:$PATH
```

Install generators:

```sh
go install github.com/twitchtv/twirp/protoc-gen-twirp
go install google.golang.org/protobuf/cmd/protoc-gen-go
```


### Old Twirp versions (v5, v7) depend on Protobuf APIv1


Older versions of Twirp require Protobuf APIv1 instead of APIv2 (See [Version Compatibility](version_matrix.md)), that has a different module name:

```sh
go get github.com/twitchtv/twirp/protoc-gen-twirp@v7.2.0
go get github.com/golang/protobuf/protoc-gen-go@1.5.2
```

### Generate code

Try the `protoc` compiler with the flags  `--twirp_out` and `--go_out` to see if it is able to generate the `twirp.go` and `.pb.go` files. See [protobuf docs](https://developers.google.com/protocol-buffers/docs/reference/go-generated) for details on how to use the protoc compiler with `--go-out`, the Twirp flag `--twirp_out` supports the same parameters (see [Generator Flags](command_line.md) for more options).

An example call with default parameters to generate code for `rpc/haberdasher/service.proto`:

```sh
protoc --go_out=. --twirp_out=. rpc/haberdasher/service.proto
```

