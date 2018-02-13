---
id: "errors"
title: "Errors"
sidebar_label: "Errors"
---

You probably noticed that all methods on a Twirp-made interface return `(...,
error)`.

Twirp clients always return errors that can be cast to `twirp.Error`. Even
transport-level errors will be `twirp.Error`s.

Twirp server implementations can return regular errors too, but those
will be wrapped with `twirp.InternalErrorWith(err)`, so they are also
`twirp.Error` values when received by the clients.

Check the [Errors Spec](spec_v5.md) for more information on error
codes and the wire protocol.

Also don't be afraid to open the [source code](https://github.com/twitchtv/twirp/blob/master/errors.go) 
for details, it is pretty straightforward.

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

Each error code is defined by a constant in the `twirp` package:

| twirp.ErrorCode    | JSON/String         |  HTTP status code
| ------------------ | ------------------- | ------------------
| Canceled           | canceled            | 408 RequestTimeout
| Unknown            | unknown             | 500 Internal Server Error
| InvalidArgument    | invalid_argument    | 400 BadRequest
| DeadlineExceeded   | deadline_exceeded   | 408 RequestTimeout
| NotFound           | not_found           | 404 Not Found
| BadRoute           | bad_route           | 404 Not Found
| AlreadyExists      | already_exists      | 409 Conflict
| PermissionDenied   | permission_denied   | 403 Forbidden
| Unauthenticated    | unauthenticated     | 401 Unauthorized
| ResourceExhausted  | resource_exhausted  | 403 Forbidden
| FailedPrecondition | failed_precondition | 412 Precondition Failed
| Aborted            | aborted             | 409 Conflict
| OutOfRange         | out_of_range        | 400 Bad Request
| Unimplemented      | unimplemented       | 501 Not Implemented
| Internal           | internal            | 500 Internal Server Error
| Unavailable        | unavailable         | 503 Service Unavailable
| DataLoss           | dataloss            | 500 Internal Server Error

For more information on each code, see the [Errors Spec](spec_v5.md).

### HTTP Errors from Intermediary Proxies

It is also possible for Twirp Clients to receive HTTP responses with non-200 status
codes but without an expected error message. For example, proxies or load balancers 
might return a "503 Service Temporarily Unavailable" body, which cannot be 
deserialized into a Twirp error.

In these cases, generated Go clients will return twirp.Errors with a code which 
depends upon the HTTP status of the invalid response:

| HTTP status code         |  Twirp Error Code
| ------------------------ | ------------------
| 3xx (redirects)          | Internal
| 400 Bad Request          | Internal
| 401 Unauthorized         | Unauthenticated
| 403 Forbidden            | PermissionDenied
| 404 Not Found            | BadRoute
| 429 Too Many Requests    | Unavailable
| 502 Bad Gateway          | Unavailable
| 503 Service Unavailable  | Unavailable
| 504 Gateway Timeout      | Unavailable
| ... other                | Unknown

Additional metadata is added to make it easy to identify intermediary errors:

* `"http_error_from_intermediary": "true"`
* `"status_code": string` (original status code on the HTTP response, e.g. `"500"`).
* `"body": string` (original non-Twirp error response as string).
* `"location": url-string` (only on 3xx reponses, matching the `Location` header).

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

Error metadata can only have string values. This is to simplify error parsing by clients.
If your service requires errors with complex metadata, you should consider adding client
wrappers on top of the auto-generated clients, or just include business-logic errors as
part of the Protobuf messages (add an error field to proto messages).

