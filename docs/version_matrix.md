---
id: version_matrix
title: Version Compatibility
sidebar_label: Version Compatibility
---

## Compatibilty Matrix

This table is not exaustive (see [Releases](https://github.com/twitchtv/twirp/releases) for details), but it contains a general sense of what runtime and protobuf library is needed for code generated on different versions.


| Twirp Generator  | Twirp Runtime | Proto Generator and Runtime | Twirp Spec |
| ---------------- |---------------| --------------------------- | ---------- |
| v8               | v7.1+         | APIV2                       | V5, V7     |
| v7.1             | v7.1+         | APIV1                       | V5, V7     |
| v7.0             | v7.0+         | APIV1                       | V5, V7     |
| v5.11            | v5.11+        | APIV1                       | V5, V7     |
| v5.10            | v5.10+        | APIV1                       | V5, V7     |
| v5.8             | v5.8+         | APIV1                       | V5, V7     |
| v5               | v5+           | APIV1                       | V5, V7     |

The Twirp library works with older versions of generated code (for the most part). This means that upgrading Twirp usually involves updating the library first, then re-generating code.


### Go Twirp and Protobuf

Both Twirp and Protobuf have runtime libraries and code generators. The generated code can have incompatibility issues with different versions of the library.

Twirp (https://github.com/twitchtv/twirp):

 * Twirp Runtime: `github.com/twitchtv/twirp`. Is the Go library with shared types like `twirp.Error` and `twirp.ServerOptions`)
 * Twirp Generator: `github.com/twitchtv/twirp/protoc-gen-twirp`. Generates Go code with the `.twirp.go` file extension, with Twirp clients and servers.

Protobuf APIV2 (new https://github.com/protocolbuffers/protobuf):

 * Proto Runtime: `google.golang.org/protobuf/proto`. Is the Proto library used to serialize Protobuf and JSON messages over the network. The new version (APIV2) is used by new versions of Twirp (v8+). The older version (APIV1) has a different import path (`github.com/golang/protobuf/proto`) and is used by older versions of Twirp (v5 and v7).
 * Proto Generator: `google.golang.org/protobuf/cmd/protoc-gen-go`. Generates Go code with the `.pb.go` file extension, with Protobuf message types.

Protobuf APIV1 (old https://github.com/golang/protobuf):

  * Proto Runtime: `github.com/golang/protobuf/proto`. Up to version 1.5.2, after that it changes the import path to APIV2.
  * Proto Generator: `google.golang.org/protobuf/cmd/protoc-gen-go`. Up to version 1.5.2, after that it changes the import path to APIV2.


### Protocol Spec Compatibility

The [Twirp Spec Protocol](https://twitchtv.github.io/twirp/docs/spec_v7.html) is the main point of compatibility for Twirp clients and services, across different versions and implementations in different languages. The Spec was first released as v5, and later updated to V7.

Golang versions of the runtime library and generator labeled with `v5.x.x` are all compliant with the V5 spec and also with the V7 spec (V7 is backwards compatible). Any old service implementing V5 also works with the V7 spec. See [V7 release notes](https://github.com/twitchtv/twirp/releases/tag/v7.0.0) for compatibility details and upgrade instructions.

Golang versions of the runtime library and generator labeled with `v7.x.x` and above (`v8+`), are compliant with the V7 spec.


