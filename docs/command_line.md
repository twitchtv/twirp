---
id: "command_line"
title: "Command line parameters"
sidebar_label: "Command line parameters"
---

In general, Twirp's Go generator shouldn't need command line parameters. There
are some complex cases, though, where they are the only way to get things done,
particularly when setting the import path to be used in generated code.

# How to pass command line parameters

Command line parameters are passed to the Twirp generator, `protoc-gen-twirp`,
by specifying them in the `--twirp_out` argument. The parameters are key-values,
separated by `,` characters, and you the parameter list is terminated with a `:` character.

So, for example, `--twirp_out=k1=v1,k2=v2,k3=v3:.` would pass `k1=v1`, `k2=v2`,
and `k3=v3` to twirp.

# Modifying imports

When working with multiple proto files that use import statements,
`protoc-gen-twirp` uses the `option go_package` field in the `.proto` files to
determine the import paths for imported message types. Usually, this is
sufficient, but in some complex setups, you need to be able to directly override
import lines.

You should usually set import paths by using `option go_package` in your .proto
files. A line like this:

```protobuf
option go_package = "github.com/twitchtv/thisisanexample";
```

will set things up properly. But if a file needs to be imported at different
paths for different users, you might need to resort to command-line parameters.


This behavior can be customized by using two different command line parameters:

* `import_prefix`, which prefixes all generated import paths with something.
* `go_import_mapping`, which lets you set an explicit mapping of import paths to
  use for particular .proto files.

## Import prefix parameter

The `import_prefix` parameter can be passed to `--twirp_out` in order to prefix
the generated import path with something.

```sh
$ PROTO_SRC_PATH=./
$ IMPORT_PREFIX="github.com/example/rpc/haberdasher"
$ protoc \
  --proto_path=$PROTO_SRC_PATH \
  --twirp_out=import_prefix=$IMPORT_PREFIX:$PROTO_SRC_PATH \
  --go_out=import_prefix=$IMPORT_PREFIX:$PROTO_SRC_PATH \
  $PROTO_SRC_PATH/rpc/haberdasher/service.proto
```

## Import mapping parameter

The import mapping parameter can be passed multiple times to `--twirp_out` in
order to substitute the import path for a given proto file with something else.
By passing the parameter multiple times you can build up a map of proto file to
import path inside the generator.

This parameter should be used when one of your proto files `import`s a proto
file from another package and you're not generating your code at the
`$GOPATH/src` root.

There are two ways to provide this parameter to `--twirp_out`:

### As provided to `protoc-gen-go`

Just like `proto-gen-go`, you can use a shorthand, formatted as: `M<proto
filename>=<go import path>`. For example, you could tell `protoc-gen-twirp` that
`rpcutil/empty.proto` can be found at `github.com/example/rpcutil` by using
`Mrpcutil/empty.proto=github.com/example/rpcutil`. Here's a full example:

```sh
$ PROTO_SRC_PATH=./
$ IMPORT_MAPPING="rpcutil/empty.proto=github.com/example/rpcutil"
$ protoc \
  --proto_path=$PROTO_SRC_PATH \
  --twirp_out=M$IMPORT_MAPPING:$PROTO_SRC_PATH \
  --go_out=M$IMPORT_MAPPING:$PROTO_SRC_PATH \
  $PROTO_SRC_PATH/rpc/haberdasher/service.proto
```

### Using the `go_import_mapping@` prefix

This is exactly the same as the previous method; it's just a little more verbose
and a little clearer. The format is `go_import_mapping@<proto filename>=<go
import path>`. For example, you could tell `protoc-gen-twirp` that
`rpcutil/empty.proto` can be found at `github.com/example/rpcutil` by using
`go_import_mapping@rpcutil/empty.proto=github.com/example/rpcutil`. Here's a
full example:

```sh
$ IMPORT_MAPPING="rpcutil/empty.proto=github.com/example/rpcutil"
$ protoc \
  --proto_path=$PROTO_SRC_PATH \
  --twirp_out=go_import_mapping@$IMPORT_MAPPING:$PROTO_SRC_PATH \
  --go_out=M$IMPORT_MAPPING:$PROTO_SRC_PATH \
  $PROTO_SRC_PATH/rpc/haberdasher/service.proto
```

Note: this is a `protoc-gen-twirp` flavor of the parameter. `protoc-gen-go` does
not support this prefix.
