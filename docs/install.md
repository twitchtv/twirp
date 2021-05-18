---
id: install
title: Installing Twirp
sidebar_label: Installation
---

### Prerequisites

 * [Go](https://golang.org/): any one of the three latest major [releases of Go](https://golang.org/doc/devel/release.html). For installation instructions, see Go’s [Getting Started](https://golang.org/doc/install) guide.
 * [Protocol buffer](https://developers.google.com/protocol-buffers) compiler, `protoc` [version 3](https://developers.google.com/protocol-buffers/docs/proto3). For installation instructions, see [Protocol Buffer Compiler Installation](https://grpc.io/docs/protoc-installation/).


### Track Generator Versions

You should track the Twirp and Protobuf generator versions like any other go-based tool (e.g. `stringer`) to ensure it always generates the same files. The currently recommended approach is to track the tool's version in your module's `go.mod` file (See ["Go Modules by Example" walkthrough](https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md)).

Add a `tools.go` file to your module that includes import statements for the tools of interest:

```go
// +build tools

package tools

import (
        _ "google.golang.org/protobuf/cmd/protoc-gen-go"
        _ "github.com/twitchtv/twirp/protoc-gen-twirp"
)
```

### Install Protobuf and Twirp Generators

Set `GOBIN` (see [go help environment](https://golang.org/cmd/go/#hdr-Environment_variables)) to define where the tool dependencies will be installed. A good idea is to have a git-ignored `/bin` folder to install tools for your project:

```sh
export GOBIN=$PWD/bin
```

Make sure that the installed packages are in your `$PATH`, so the `protoc` compiler can find them. You might need to add GOBIN to your PATH:

```sh
export PATH=$GOBIN:$PATH
```

Download and install generators:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install github.com/twitchtv/twirp/protoc-gen-twirp
```


### Old versions (v5, v7) with Protobuf APIv1


Older versions of Twirp require Protobuf APIv1 instead of APIv2 (See [Version Compatibility](version_matrix.md)). For example:

```sh
go get github.com/golang/protobuf/protoc-gen-go@1.5.2
go get github.com/twitchtv/twirp/protoc-gen-twirp@v7.2.0
```

### Generate code

Try the `protoc` compiler with the `--go_out` and `--twirp_out` options to see if it is able to generate the `.pb.go` and `twirp.go` files. See https://developers.google.com/protocol-buffers/docs/reference/go-generated for details on how to use the protoc compiler with `--go-out`. The Twirp flag `--twirp_out` supports the same parameters. See [Generator Flags](command_line.md) for more options.

For example, to invoke the protoc compiler with default parameters to generate code for the proto named `rpc/haberdasher/service.proto`

```sh
protoc --twirp_out=. --go_out=. rpc/haberdasher/service.proto
```

