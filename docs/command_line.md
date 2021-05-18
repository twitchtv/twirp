---
id: "command_line"
title: "Generator Flags for the Protoc Compiler"
sidebar_label: "Generator Flags"
---


### Protoc Compiler Invocation

Make sure to [Install Twirp and Protobuf](install.md) generators first.

The `protoc` compiler uses the flag  `--twirp_out` to generate `twirp.go` files, and `--go_out` to generate `.pb.go` files. See [Protobuf Docs](https://developers.google.com/protocol-buffers/docs/reference/go-generated) for details on how to use the protoc compiler with `--go-out`.

The Twirp flag `--twirp_out` supports the same parameters.

For example, invoke the protoc compiler with default parameters to generate code for `rpc/haberdasher/service.proto`:

```sh
protoc --go_out=. --twirp_out=. rpc/haberdasher/service.proto
```


### Modifying imports

The protoc compiler invocation can include optional flags to set the import path to be used in generated code: `--go_opt` and `--twirp_opt`.

When working with multiple proto files that use import statements, `protoc-gen-twirp` uses the `option go_package` field in the `.proto` files to determine the import paths for imported message types. For example:

```protobuf
option go_package = "github.com/twitchtv/thisisanexample";
```

The `import_prefix` parameter can be passed to `--twirp_out` in order to prefix the generated import path with something.

```sh
$ PROTO_SRC_PATH=./
$ IMPORT_PREFIX="github.com/example/rpc/haberdasher"
$ protoc \
  --proto_path=$PROTO_SRC_PATH \
  --twirp_out=import_prefix=$IMPORT_PREFIX:$PROTO_SRC_PATH \
  --go_out=import_prefix=$IMPORT_PREFIX:$PROTO_SRC_PATH \
  $PROTO_SRC_PATH/rpc/haberdasher/service.proto
```

Alternatively, an import mapping parameter can be passed multiple times to `--twirp_out` in
order to substitute the import path for a given proto file with something else.
By passing the parameter multiple times you can build up a map of proto file to
import path inside the generator.

This parameter should be used when one of your proto files imports a proto
file from another package and you're not generating your code at the
`$GOPATH/src` root.

For example, you could tell `protoc-gen-twirp` that
`rpcutil/empty.proto` can be found at `github.com/example/rpcutil` by using
`Mrpcutil/empty.proto=github.com/example/rpcutil`:

```sh
$ PROTO_SRC_PATH=./
$ IMPORT_MAPPING="rpcutil/empty.proto=github.com/example/rpcutil"
$ protoc \
  --proto_path=$PROTO_SRC_PATH \
  --twirp_out=M$IMPORT_MAPPING:$PROTO_SRC_PATH \
  --go_out=M$IMPORT_MAPPING:$PROTO_SRC_PATH \
  $PROTO_SRC_PATH/rpc/haberdasher/service.proto
```
