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

```
POST <baseURL>/<prefix>/<package>.<Service>/<Method>
```

Where:

 * The `<baseURL>` is the URL `scheme` and `authority` where the service is located. For example, "https://example.com".
 * The `<prefix>` is "/twirp" by default, but it is optional and can be configured on services and clients through constructor options (`twirp.WithServicePrefix` and `twirp.WithClientPrefix`). The prefix can be empty ("") or have multiple components (e.g. "/my/custom/prefix").
 * The `<package>`, `<Service>` and `<Method>` names are the same values used in the `.proto` file where the service was defined.

Examples of valid Twirp routes:

```
POST https://example.com/twirp/twirp.example.haberdasher.Haberdasher/MakeHat
POST https://example.com/twirp/mypackage.MyService/MyMethod
POST https://example.com/my/custom/prefix/mypackage.MyService/MyMethod
```

More details on the [protocol specification](spec_v5.md).

### Naming Stype Guide

It is higly recommended that the `<Service>` and `<Method>` names are CamelCased, as recommended by the [Protocol Buffers Style Guide](https://developers.google.com/protocol-buffers/docs/style#services)).

The [official Go implementation](https://github.com/twitchtv/twirp) differs in behavior from what is described in the [specification](https://twitchtv.github.io/twirp/docs/spec_v5.html). It modifies the service and method names to be CamelCase (with an initial capital) instead of using the exact names specified in the protobuf definition. This means that the URL paths generated for Go clients and servers may differ from paths generated for other language implementations. This issue is discussed in [#244](https://github.com/twitchtv/twirp/issues/244).

```
// Good
POST /twirp/mypackage.MyService/MyMethod

// Potentially problematic
POST /twirp/mypackage.my_service/my_method
```

### Content-Type Header (json or protobuf)

The `Content-Type` header is required and must be either `application/json` or
`application/protobuf`. JSON is easier for debugging (particularly when making
requests with cURL), but Protobuf is better in almost every other way. Please
use Protobuf in real code. See
[Protobuf and JSON](https://github.com/twitchtv/twirp/wiki/Protobuf-and-JSON)
for more details.

### JSON serialization

The JSON format should match the
[official spec](https://developers.google.com/protocol-buffers/docs/proto3#json)'s
rules for JSON serialization. In a nutshell: names are `camelCased`, _all_
fields must be set, _no_ extra fields may be set, and `null` means "I want to
leave this field blank".

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

## Making requests on the command line with cURL

See [cURL](cURL.md)
