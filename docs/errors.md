---
id: "errors"
title: "Errors"
sidebar_label: "Errors"
---

Twirp clients always return errors that can be cast to `twirp.Error`.
Transport-level errors are wrapped as `twirp.Error`.

Twirp server implementations can return regular errors too, but those
will be wrapped with `twirp.InternalErrorWith(err)`, so they are also
`twirp.Error` values when received by the clients.

Check the [Errors Spec](spec_v7.md) and the [source code](https://github.com/twitchtv/twirp/blob/master/errors.go)
for more information on error codes and the wire protocol.

### twirp.Error interface

```go
type Error interface {
    Code() ErrorCode        // identifies a valid error type
    Msg() string            // free-form human-readable message

    WithMeta(key string, val string) Error // set metadata
    Meta(key string) string                // get metadata value
    MetaMap() map[string]string            // see all metadata

    Error() string // as an error returns "twirp error <Code>: <Msg>"
}
```

### Error Codes

Error codes are defined by a constant in the `twirp` package.
Check the [Errors Spec](spec_v7.md) and the [source code](https://github.com/twitchtv/twirp/blob/master/errors.go)
for more information on error codes and the wire protocol.

### HTTP Errors from Intermediary Proxies

Twirp Clients may receive HTTP responses with non-200 status
from different sources like proxies or load balancers. For example,
a "503 Service Temporarily Unavailable" body, which cannot be
deserialized into a Twirp error.

In those cases, generated Go clients will return `twirp.Error` with a code
depending on the HTTP status of the invalid response:

| HTTP status code         |  Twirp Error Code
| ------------------------ | ------------------
| 3xx (redirects)          | Internal
| 400 Bad Request          | Internal
| 401 Unauthorized         | Unauthenticated
| 403 Forbidden            | PermissionDenied
| 404 Not Found            | BadRoute
| 429 Too Many Requests    | ResourceExhausted
| 502 Bad Gateway          | Unavailable
| 503 Service Unavailable  | Unavailable
| 504 Gateway Timeout      | Unavailable
| ... other                | Unknown

Additional metadata is added to make it easy to identify intermediary errors:

* `"http_error_from_intermediary": "true"`
* `"status_code": string` (original status code on the HTTP response, e.g. `"500"`).
* `"body": string` (original non-Twirp error response as string).
* `"location": url-string` (only on 3xx responses, matching the `Location` header).

### Metadata

Arbitrary string metadata can be added to any error. For example, a service may return this:

```go
if unavailable {
    twerr := twirp.NewError(twirp.Unavailable, "taking a nap ...")
    twerr = twerr.WithMeta("retryable", "true")
    twerr = twerr.WithMeta("retry_after", "15s")
    return nil, twerr
}
```

The metadata is available on the client:

```go
if twerr.Code() == twirp.Unavailable {
    if twerr.Meta("retry_after") != "" {
        // ... retry after twerr.Meta("retry_after")
    }
}
```

Error metadata can only have string values. This is to simplify error parsing by clients.
If your service requires errors with complex metadata, you should consider adding client
wrappers on top of the auto-generated clients, or just include business-logic errors as
part of the Protobuf messages (add an error field to proto messages).

### Writing HTTP Errors outside Twirp services

Twirp services can be [muxed with other HTTP services](mux.md). For consistent responses
and error codes _outside_ Twirp servers, such as http middlewares, you can call `twirp.WriteError`.

The error is expected to satisfy a `twirp.Error`, otherwise it is wrapped with `twirp.InternalError`.

Usage:

```go
rpc.WriteError(w, twirp.NewError(twirp.Unauthenticated, "invalid token"))
```

To simplify `twirp.Error` composition, a few constructors are available, such as `NotFoundError`
and `RequiredArgumentError`. See [docs](https://godoc.org/github.com/twitchtv/twirp#Error).

With constructor:

```go
rpc.WriteError(w, twirp.RequiredArgumentError("user_id"))
```
