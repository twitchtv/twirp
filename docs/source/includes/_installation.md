# Installation

Twirp works for any project written in Go1.7+. You'll need these things, too:

* `protoc`, the protobuf compiler
* `protoc-gen-go`, the go plugin for `protoc`
* `protoc-gen-twirp`, the twirp plugin for `protoc`

## Install protoc

> Run the following command to install `protoc` using homebrew:

```bash
$ brew install protobuf
```

[Install Protocol Buffers v3](https://developers.google.com/protocol-buffers/docs/gotutorial)
, the protoc compiler that is used to auto-generate code. The simplest way to do this is to download pre-compiled binaries for your platform from here:
[https://github.com/google/protobuf/releases](https://github.com/google/protobuf/releases).

## Install protoc-gen-go

> Download and install `protoc-gen-go` using the normal Go tools:

```bash
$ go get -u github.com/golang/protobuf/protoc-gen-go
```

The normal Go tools will install `protoc-gen-go` in `$GOBIN`, defaulting to `$GOPATH/bin`.


<div class="clear"></div>

> Add `$GOBIN` to your path if you haven't already:

```bash
$ export PATH=$PATH:$GOPATH/bin
```

It must be in your `$PATH` for the protocol compiler, `protoc`, to find it, so you might need to explicitly add it to your path:

## Install protoc-gen-twirp

> Install the Twirp compiler plugin `protoc-gen-twirp`:

```bash
$ go get -u github.com/twitchtv/twirp/protoc-gen-twirp
```

Installing `protoc-gen-twirp` will add the executable to your `$GOBIN` as well. When `protoc` is run, it will look for plugins such as `protoc-gen-twirp` and `protoc-gen-go` in your `$PATH`.