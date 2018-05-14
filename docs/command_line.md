---
id: "command_line"
title: "Command line"
sidebar_label: "Command line"
---

When generating code, by default Twirp will use the name of the directory as the package name.

This behavior can be customized by using two different command line parameters:

* import prefix
* import mapping

## Import prefix parameter

The `import_prefix` parameter can be passed to `--twirp_out` in order to prefix the generated import path with something.

```sh
$ PROTO_SRC_PATH=./
$ IMPORT_PREFIX="github.com/example/rpc/haberdasher"
$ protoc --proto_path=$PROTO_SRC_PATH --twirp_out=import_prefix=$IMPORT_PREFIX:$PROTO_SRC_PATH --go_out=import_prefix=$IMPORT_PREFIX:$PROTO_SRC_PATH $PROTO_SRC_PATH/rpc/haberdasher/service.proto
```

## Import mapping parameter

The import mapping parameter can be passed multiple times to `--twirp_out` in order to substitute the import path for a given proto file with something else. By passing the parameter multiple times you can build up a map of proto file to import path inside the generator.

This parameter should be used when one of your proto files `import`s a proto file from another package and you're not generating your code at the `$GOPATH/src` root.

There are two ways to provide this parameter to `--twirp_out`:

### As provided to `protoc-gen-go`

```sh
$ PROTO_SRC_PATH=./
$ IMPORT_MAPPING="rpcutil/empty.proto=github.com/example/rpcutil"
$ protoc --proto_path=$PROTO_SRC_PATH --twirp_out=M$IMPORT_MAPPING:$PROTO_SRC_PATH --go_out=M$IMPORT_MAPPING:$PROTO_SRC_PATH $PROTO_SRC_PATH/rpc/haberdasher/service.proto
```

### Using the `go_import_mapping@` prefix

This is exactly the same as the previous way. It's simply more descriptive.

```sh
$ IMPORT_MAPPING="rpcutil/empty.proto=github.com/example/rpcutil"
$ protoc --proto_path=$PROTO_SRC_PATH --twirp_out=go_import_mapping@$IMPORT_MAPPING:$PROTO_SRC_PATH --go_out=M$IMPORT_MAPPING:$PROTO_SRC_PATH $PROTO_SRC_PATH/rpc/haberdasher/service.proto
```

Note: this is a `protoc-gen-twirp` flavor of the parameter. `protoc-gen-go` does not support this prefix.
