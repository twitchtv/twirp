---
id: "errors"
title: "Errors"
sidebar_label: "Errors"
---

A Twirp error has:

 * **code**: identifies the type of error.
 * **msg**: free-form message with detailed information about the error. It is meant for humans, to assist with debugging. Programs should not try to parse the error message.
 * **meta**: (optional) key-value pairs with arbitrary string metadata.

## Error Codes

Valid Twirp error codes (HTTP status):

 * `internal` (500)
 * `not_found` (404)
 * `invalid_argument` (400)
 * `unauthenticated` (401)
 * `permission_denied` (403)
 * `already_exists` (409)
 * ... more on the [Errors Spec](spec_v7.md#error-codes)

To map a [twirp.ErrorCode](https://pkg.go.dev/github.com/twitchtv/twirp#ErrorCode) into the equivalent HTTP status, use the helper [twirp.ServerHTTPStatusFromErrorCode](https://pkg.go.dev/github.com/twitchtv/twirp#ServerHTTPStatusFromErrorCode)).

## Overview

A Twirp endpoint returns a [twirp.Error](https://pkg.go.dev/github.com/twitchtv/twirp#Error). For example, a "Permission

```go
func (s *Server) Foo(ctx context.Context, req *pb.FooRequest) (*pb.FooResp, error) {
    return nil, twirp.PermissionDenied.Error("this door is closed")
}
```

Twirp serializes the response as a JSON with `code` and `msg` keys:

```json
// HTTP status: 403
{
  "code": "permission_denied",
  "msg": "this door is closed"
}
```

The auto-generated client de-serializes and returns the same Twirp error:

```go
resp, err := client.Foo(ctx, req)
if twerr, ok := err.(twirp.Error); ok {
    twerr.Code() // => twirp.PermissionDenied
    twerr.Msg() //=> "this door is closed"
}
```

## Server Side: Returning Error Responses

A Twirp endpoint may return an error. If the error value implements the  interface, it will be serialized and received by the client with the exact same `code`, `msg` and `meta` properties.

The `twirp` package provides error constructors for each code. For example, to build an internal error: `twirp.Internal.Error("oops")`. There is also a generic constructor [twirp.NewError](https://pkg.go.dev/github.com/twitchtv/twirp#NewError). Anything that implements the `twirp.Error` interface counts as a Twirp error. Check the [errors.go file for details](https://github.com/twitchtv/twirp/blob/main/errors.go).

Example of an endpoint returning Twirp errors:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    // Validation errors
    if req.UserId == "" {
        return nil, twirp.InvalidArgument.Error("user_id is required")
    }

    // Perform some operation
    user, err := s.DB.FindByID(ctx, req.UserID)
    if errors.Is(err, DB_NOT_FOUND) {
        return nil, twirp.NotFound.Error("user not found")
    }
    if err != nil {
        return nil, twirp.Internal.Errorf("DB error: %w", err)
    }

    // Success
    return &pb.FindUserResp{
        Login: user.Login,
    }, nil
}
```

If the endpoint returns a vanilla (non-twirp) error, it will be automatically wrapped as an **internal** error with [twirp.InternalErrorWith(err)](https://pkg.go.dev/github.com/twitchtv/twirp#InternalErrorWith)).

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    err := errors.New("oops")
    return nil, err
}
```

Using the wrapper explicitly is equivalent; the client will receive the same internal error.

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    err := errors.New("oops")
    return nil, twirp.InternalErrorWith(err)
}
```

And that is equivalent to building the internal error like this:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    err := errors.New("oops")

	return twirp.Internal.Errorf("%w", err).
        WithMeta("cause", fmt.Sprintf("%T", err))
}
```

#### Middleware, outside Twirp endpoints

Twirp services can be [muxed with other HTTP services](mux.md). For consistent responses and error codes _outside_ Twirp servers, such as HTTP middleware, you can call [twirp.WriteError](https://pkg.go.dev/github.com/twitchtv/twirp#WriteError).

```go
twirp.WriteError(responseWriter, twirp.Unauthenticated.Error("invalid token"))
```


## Client Side: Handling Error Responses

Twirp clients return errors that can always be cast to the `twirp.Error` interface. Unpack the error type to access the `Code()`, `Msg()` and `Meta(key)` properties:

```go
resp, err := client.FindUser(ctx, req)
if err != nil {
    if twerr, ok := err.(twirp.Error); ok {
        if twerr.Code() == twirp.NotFound {
            fmt.Println("not found")
        }
    }
    fmt.Printf("internal: %s", err)
}
```

You can also use [errors.Is](https://pkg.go.dev/errors#Is) and [errors.As](https://pkg.go.dev/errors#As) to check and unwrap Twirp errors:

```go
resp, err := client.MakeHat(ctx, req)
var twerr twirp.Error
if errors.As(err, &twerr) {
    if twerr.Code() == twirp.NotFound {
        fmt.Println("not found")
    }
} else if err != nil {
    fmt.Printf("internal: %s", err)
}
```

Transport-level errors (e.g. connection issues) are returned as internal errors. If desired, the original client-side network error can be unwrapped:

```go
resp, err := client.MakeHat(ctx, req)
var twerr twirp.Error
if errors.As(err, &twerr) {
    if twerr.Code() == twirp.Internal {
        if transportErr := errors.Unwrap(twerr); transportErr != nil {
            // transportErr could be something like an HTTP connection error
        }
    }
}
```

### HTTP Errors from Intermediary Proxies

Twirp Clients may receive HTTP responses with non-200 status
from different sources like proxies or load balancers. For example,
a "503 Service Temporarily Unavailable" body, which cannot be
deserialized into a Twirp error.

In those cases, generated Go clients will try to best-guess the equivalent Twirp error depending on the HTTP status of the invalid response:

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


## Metadata

In addition to `code` and `msg`, Twirp errors can optionally include arbitrary string metadata in the `meta` field.

For example, some server code could return an error like this:

```go
if unavailable {
    return nil, twirp.Unavailable.Error("taking a nap ...").
        WithMeta("retryable", "true").
        WithMeta("retry_after", "15s")
}
```

Twirp serializes the response as a JSON with the additional `meta` field:

```json
// HTTP status code: 503
{
  "code": "unavailable",
  "msg": "taking a nap ...",
  "meta": {
    "retryable": "true",
    "retry_after": "15s"
  }
}
```

Metadata is available on the client using the `.Meta(key)` accessor:

```go
if twerr.Code() == twirp.Unavailable {
    if twerr.Meta("retryable") == "true" {
        fmt.Printf("retry after %s", twerr.Meta("retry_after"))
    }
}
```

Error metadata can only have string values. This is to simplify error parsing by client implementations in multiple platforms. If your service requires errors with complex shapes, consider adding client wrappers on top of the auto-generated clients, or include specific business-logic errors on the Protobuf messages (as part of success responses).
