---
id: "routing"
title: "Routing and Serialization"
sidebar_label: "Routing and Serialization"
---

Routing and Serialization is handled by Twirp. All you really need to know is
that "it just works". However you may find this interesting and useful for
debugging and advanced configuration.

### HTTP Routes

Twirp works over HTTP 1.1; all RPC methods map to routes that follow the format:

```plaintext
POST [<prefix>]/[<package>.]<Service>/<Method>
```

Where:

 * The `<prefix>` is "/twirp" by default, but it is optional and can be configured with other paths like "/my/custom/prefix" (see `twirp.WithServerPathPrefix`).
 * The `<package>`, `<Service>` and `<Method>` names are the same values used in the `.proto` file where the service was defined.

For example, to make a hat with the [Haberdasher service](example.md):

```plaintext
POST /twirp/twirp.example.haberdasher.Haberdasher/MakeHat
```

More details on the [protocol specification](spec_v7.md).

#### Naming Style

For maximum compatibility, please follow the [Protocol Buffers Style Guide](https://developers.google.com/protocol-buffers/docs/style#services). In particular, the `<Service>` and `<Method>` names should be CamelCased (with an initial capital). This will ensure cross-language compatibility and prevent name collisions (e.g. `myMethod` and `my_method` would both map to `MyMethod`, causing a compile time error in some languages like Go).

### Content-Type Header (json or protobuf)

The `Content-Type` header is required and must be either `application/json` or
`application/protobuf`. JSON is easier for debugging (particularly when making
requests with cURL), but Protobuf is better in almost every other way. Please
use Protobuf in real code. See
[Protobuf and JSON](https://github.com/twitchtv/twirp/wiki/Protobuf-and-JSON)
for more details.

### JSON serialization

The JSON format generally matches the
[official spec](https://developers.google.com/protocol-buffers/docs/proto3#json)'s
rules for JSON serialization.  In a nutshell: _all_
fields must be set, _no_ extra fields may be set, and `null` means "I want to
leave this field blank".
One exception to this is that names match the proto names, this is considered more predictable to those writing custom JSON clients.

### Error responses

Errors returned by Twirp servers use non-200 HTTP status codes and always have
JSON-encoded bodies (even if the request was Protobuf-encoded). The body JSON
has three fields `{type, msg, meta}`. For example:

```
POST /twirp/twirp.example.haberdasher.Haberdasher/INVALIDROUTE

404 Not Found
{
    "type": "bad_route",
    "msg": "no handler for path /twirp/twirp.example.haberdasher.Haberdasher/INVALIDROUTE",
    "meta": {"twirp_invalid_route": "POST /twirp/twirp.example.haberdasher.Haberdasher/INVALIDROUTE"}
}
```

