---
id: install
title: Installing Twirp
sidebar_label: Installation
---

You'll need a few things to install Twirp:
 * Go1.7+
 * The protobuf compiler `protoc`
 * Go and Twirp protoc plugins `protoc-gen-go` and `protoc-gen-twirp`

## Install protoc

[Install Protocol Buffers v3](https://developers.google.com/protocol-buffers/docs/gotutorial),
the `protoc` compiler that is used to auto-generate code. The simplest way to do
this is to download pre-compiled binaries for your platform from here:
https://github.com/google/protobuf/releases

It is also available in MacOS through Homebrew:

```sh
$ brew install protobuf
```

## Get protoc-gen-go and protoc-gen-twirp plugins

### With retool

We recommend using [retool](https://github.com/twitchtv/retool) to manage go
tools like commands and linters:

```sh
$ go get github.com/twitchtv/retool
```

Install the plugins into your project's `_tools` folder:
```sh
$ retool add github.com/golang/protobuf/protoc-gen-go master
$ retool add github.com/twitchtv/twirp/protoc-gen-twirp master
```

This will make it easier to manage and update versions without causing problems
to other project collaborators.

If the plugins were installed with retool, when run the `protoc` command make
sure to prefix with `retool do`, for example:

```sh
$ retool do protoc --proto_path=$GOPATH/src:. --twirp_out=. --go_out=. ./rpc/haberdasher/service.proto
```

### With go get

Download and install `protoc-gen-go` using the normal Go tools:

```sh
$ go get -u github.com/golang/protobuf/protoc-gen-go
$ go get -u github.com/twitchtv/twirp/protoc-gen-twirp
```

The normal Go tools will install `protoc-gen-go` in `$GOBIN`, defaulting to
`$GOPATH/bin`. It must be in your `$PATH` for the protocol compiler, `protoc`,
to find it, so you might need to explicitly add it to your path:

```sh
$ export PATH=$PATH:$GOPATH/bin
```

## Updating Twirp ##

Twirp releases are tagged with semantic versioning and releases are managed by
Github. See the [releases](https://github.com/twitchtv/twirp/releases) page.

To stay up to date, you update `protoc-gen-twirp` and regenerate your code. If
you are using [retool](https://github.com/twitchtv/retool), that's done with

```sh
$ retool upgrade github.com/twitchtv/twirp/protoc-gen-twirp v5.2.0
```

If you're not using retool, you can also do a system-wide install with checking
out the package new version and using `go install`:

```sh
$ cd $GOPATH/src/github.com/twitchtv/twirp
$ git checkout v5.2.0
$ go install ./protoc-gen-twirp
```

With the new version of `protoc-gen-twirp`, you can re-generate code to update
your servers. Then, any of the clients of your service can update their vendored
copy of your service to get the latest version.
