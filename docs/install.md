---
id: install
title: Installing Twirp
sidebar_label: Installation
---

You'll need a few things to install Twirp:

 * [Go](https://golang.org/doc/install), Twirp supports the last 3 major versions
 * The protobuf compiler `protoc`
 * Go and Twirp protoc plugins `protoc-gen-go` and `protoc-gen-twirp`

## Install protoc

[Install Protocol Buffers v3](https://developers.google.com/protocol-buffers/docs/gotutorial),
the `protoc` compiler that is used to auto-generate code. The simplest way to do
this is to download pre-compiled binaries for your platform from here:
https://github.com/google/protobuf/releases or in MacOS `brew install protobuf`.

## Get protoc-gen-go and protoc-gen-twirp plugins

### With go get

```sh
$ go get google.golang.org/protobuf/cmd/protoc-gen-go
$ go get github.com/twitchtv/twirp/protoc-gen-twirp
```

The normal Go tools will install `protoc-gen-go` in `$GOBIN`, defaulting to
`$GOPATH/bin`. It must be in your `$PATH` for the protocol compiler, `protoc`,
to find it, so you might need to explicitly add it to your path:

```sh
$ export PATH="$PATH:$(go env GOPATH)/bin"
```

You can also add the `export` above to your `.bashrc` file and source it when needed.

## Updating Twirp

Twirp releases are tagged with semantic versioning and releases are managed by
Github. See the [releases](https://github.com/twitchtv/twirp/releases) page.

To stay up to date, you update `protoc-gen-twirp` and regenerate your code.

To upgrade you can do a system-wide install with checking
out the package new version and using `go install`:

```sh
$ cd $GOPATH/src/github.com/twitchtv/twirp
$ git checkout v5.2.0
$ go install ./protoc-gen-twirp
```

With the new version of `protoc-gen-twirp`, you can re-generate code to update
your servers. Then, any of the clients of your service can update their vendored
copy of your service to get the latest version.
