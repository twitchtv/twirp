---
id: version_matrix
title: Version Compatibility
sidebar_label: Version Compatibility
---

## Compatibilty Matrix

Code generated with the Twirp Generator on the left, is compatible with the runtime, protobuf runtime and generated code, and spec versions on the right.

| Twirp Generator  | Twirp Runtime | Protobuf | Twirp Spec | Key feature |
| ---------------- |---------------| ---------| ---------- | ------------|
| v8               | v7.1+         | APIV2    | V7         | [Protobuf APIV2](https://github.com/twitchtv/twirp/releases/tag/v8.0.0)
| v7.1             | v7.1+         | APIV1    | V7         | [Interceptors](https://github.com/twitchtv/twirp/releases/tag/v7.1.0)
| v7.0             | v7.0+         | APIV1    | V7         | [V7 Spec and ServerOptions](https://github.com/twitchtv/twirp/releases/tag/v7.0.0)
| v5.11            | v5.10+        | APIV1    | V5, V7     | [Unwrap errors](https://github.com/twitchtv/twirp/releases/tag/v5.11.0)
| v5.10            | v5.10+        | APIV1    | V5, V7     | [ClientHooks](https://github.com/twitchtv/twirp/releases/tag/v5.10.0)
| v5.8             | v5.8+         | APIV1    | V5, V7     | [Marlformed Error](https://github.com/twitchtv/twirp/releases/tag/v5.8.0)
| v5               | v5+           | APIV1    | V5, V7     | [First Public Release](https://github.com/twitchtv/twirp/releases/tag/v5.0.0)

This table is not exaustive, see [Releases](https://github.com/twitchtv/twirp/releases) for details on each version.

### Go Twirp and Protobuf

Both Twirp and Protobuf have runtime libraries and code generators. The generated code can have incompatibility issues with different versions of the library.

Twirp (https://github.com/twitchtv/twirp):

 * Twirp Generator: `github.com/twitchtv/twirp/protoc-gen-twirp`. Generates Go code with the `.twirp.go` file extension, with Twirp clients and servers.
 * Twirp Runtime: `github.com/twitchtv/twirp`. Is the Go library with shared types like `twirp.Error` and `twirp.ServerOptions`)

Protobuf APIV2 (https://github.com/protocolbuffers/protobuf-go):

 * Proto Generator: `google.golang.org/protobuf/cmd/protoc-gen-go`. Generates Go code with the `.pb.go` file extension, with Protobuf message types.
 * Proto Runtime: `google.golang.org/protobuf/proto`. Is the Proto library used to serialize Protobuf and JSON messages over the network. The new version (APIV2) is used by new versions of Twirp (v8+). The older version (APIV1) has a different import path (`github.com/golang/protobuf/proto`) and is used by older versions of Twirp (v5 and v7).

Protobuf APIV1 (DEPRECATED: https://github.com/golang/protobuf):

  * Proto Generator: `github.com/golang/protobuf/protoc-gen-go`.
  * Proto Runtime: `github.com/golang/protobuf/proto`.


### Protocol Spec Compatibility

The [Twirp Spec Protocol](https://twitchtv.github.io/twirp/docs/spec_v7.html) is the main point of compatibility for Twirp clients and services, across different versions and implementations in different languages. The Spec was first released as v5, and later updated to V7.

Golang versions of the runtime library and generator labeled with `v5.x.x` are all compliant with the V5 spec and also with the V7 spec (V7 is backwards compatible). Any old service implementing V5 also works with the V7 spec. See [V7 release notes](https://github.com/twitchtv/twirp/releases/tag/v7.0.0) for compatibility details and upgrade instructions.

Golang versions of the runtime library and generator labeled with `v7.x.x` and above (`v8+`), are compliant with the V7 spec.


