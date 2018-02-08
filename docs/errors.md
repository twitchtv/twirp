---
id: "errors"
title: "Errors"
sidebar_label: "Errors"
---

You probably noticed that all methods on a Twirp-made interface return `(...,
error)`.

Twirp clients always return errors that can be cast to `twirp.Error`. Even
transport-level errors will be `twirp.Error`s.

Twirp server implementations can return regular errors, but if they do, they
will be wrapped with `twirp.InternalErrorWith(err)`, so they are also
`twirp.Error` values when received by the clients.

Don't be afraid to check the source code for details, it is pretty
straightforward: [errors.go](https://github.com/twitchtv/twirp/blob/master/errors.go)

### twirp.Error interface

Twirp Errors have this interface:
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

The possible values for `ErrorCode` are inspired by
[gRPC status codes](https://godoc.org/google.golang.org/grpc/codes):

```
ErrorCode            JSON/String             HTTP status code
-------------------  ----------------------  -------------------
Canceled           | "canceled"            | 408 RequestTimeout
Unknown            | "unknown"             | 500 Internal Server Error
InvalidArgument    | "invalid_argument"    | 400 BadRequest
DeadlineExceeded   | "deadline_exceeded"   | 408 RequestTimeout
NotFound           | "not_found"           | 404 Not Found
BadRoute           | "bad_route"           | 404 Not Found
AlreadyExists      | "already_exists"      | 409 Conflict
PermissionDenied   | "permission_denied"   | 403 Forbidden
Unauthenticated    | "unauthenticated"     | 401 Unauthorized
ResourceExhausted  | "resource_exhausted"  | 403 Forbidden
FailedPrecondition | "failed_precondition" | 412 Precondition Failed
Aborted            | "aborted"             | 409 Conflict
OutOfRange         | "out_of_range"        | 400 Bad Request
Unimplemented      | "unimplemented"       | 501 Not Implemented
Internal           | "internal"            | 500 Internal Server Error
Unavailable        | "unavailable"         | 503 Service Unavailable
DataLoss           | "dataloss"            | 500 Internal Server Error
NoError            | ""                    | 200 OK
```

The most common ErrorCodes are probably `InvalidArgument`, `NotFound` and `Internal`.

Documentation for each ErrorCode is in [the godoc page](https://godoc.org/github.com/twitchtv/twirp#ErrorCode).

### Metadata

You can add arbitrary string metadata to any error. For example, the service may return an error like this:

```go
if unavailable {
    twerr := twirp.NewError(twirp.Unavailable, "taking a nap ...")
    twerr = twerr.WithMeta("retryable", "true")
    twerr = twerr.WithMeta("retry_after", "15s")
    return nil, twerr
}
```

And the metadata is available on the client:

```go
if twerr.Code() == twirp.Unavailable {
    if twerr.Meta("retryable") != "" {
        // do stuff... maybe retry after twerr.Meta("retry_after")
    }
}
```

### Constructors

```go
// Generic constructor
twirp.NewError(code twirp.ErrorCode, msg string) twirp.Error

// Convenience constructors for common errors
twirp.NotFoundError(msg string) twirp.Error
twirp.InvalidArgumentError(arg, msg  string) twirp.Error
twirp.RequiredArgumentError(arg  string) twirp.Error
twirp.InternalError(msg  string) twirp.Error
twirp.InternalErrorWith(err error) twirp.Error
```

### Errors responses are JSON

Errors returned by Twirp servers use non-200 HTTP status codes and always have
JSON-encoded bodies (even if the request was Protobuf-encoded). The body JSON
has three fields {code, msg, meta}.

For example, an error returned from a Twirp method like this:

```go
twerr := twirp.NewError(twirp.PermissionDenied, "thou shall not pass")
twerr = twerr.WithMeta("target", "Balrog")
return nil, twerr
```

serializes to this JSON response body:

```json
{
    "code": "permission_denied",
    "msg": "thou shall not pass",
    "meta": {
        "target": "Balrog"
    }
}
```
