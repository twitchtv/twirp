---
id: "errors"
title: "Errors"
sidebar_label: "Errors"
---

A Twirp error has the following properties:

 * **code**: Identifies the type of error.
 * **msg**: Free-form message with detailed information about the error. This is for humans, to help with debugging. Programs should not try to parse the error message.
 * **meta**: (optional) key-value pairs with arbitrary string metadata. Useful to define subtypes under the same code, or add extra fields for the callers.

On the network, an error is represented as a JSON response with those properties. In Go, any value that implements the [twirp.Error](https://pkg.go.dev/github.com/twitchtv/twirp#Error) interface is considered a Twirp error. Other languages have different ways to represent the errors, but they always have the same properties and valid set of codes.

## Error Codes

Twirp error codes with equivalent [HTTP status](https://pkg.go.dev/github.com/twitchtv/twirp#ServerHTTPStatusFromErrorCode):

 * `internal` (500)
 * `not_found` (404)
 * `invalid_argument` (400)
 * `unauthenticated` (401)
 * `permission_denied` (403)
 * `already_exists` (409)
 * ... more on the [Errors Spec](spec_v7.md#error-codes)

## Overview

A Twirp service may implement an endpoint that returns an error. For example:

```go
func (s *Server) OpenDoor(ctx context.Context, req *pb.OpenDoorRequest) (*pb.OpenDoorResp, error) {
    return nil, twirp.PermissionDenied.Error("this door is closed")
}
```

The service HTTP response becomes be the error serialized as JSON:

```json
// HTTP status: 403
{
  "code": "permission_denied",
  "msg": "this door is closed"
}
```

Calling the endpoint from an auto-generated client will result on the same error, that can be inspected through the properties on the [twirp.Error](https://pkg.go.dev/github.com/twitchtv/twirp#Error) interface:

```go
resp, err := client.OpenDoor(ctx, req)
if twerr, ok := err.(twirp.Error); ok {
    twerr.Code() // => twirp.PermissionDenied
    twerr.Msg()  //=> "this door is closed"
}
```

## Server Side: Returning Error Responses

The `twirp` package provides a variety of error constructors. Check the [errors.go file for details](https://github.com/twitchtv/twirp/blob/main/errors.go). Some examples:

```go
// (twirp.Code).Error(msg) to build a new error from the code
twirp.Internal.Error("oops")
twirp.NotFound.Error("user not found")
twirp.InvalidArgument.Error("user_id must be alphanumeric")

// (twirp.Code).Errorf(msg, ...args) to wrap other errors
twirp.Internal.Errorf("Failed to perform operation: w%", err)

// Generic constructor
twirp.NewError(twirp.InvalidArgument, "user_id must be alphanumeric")

// Any value that implements the twirp.Error interface
myOwnTwirpErrImpl{code: twirp.NotFound}
```

Example of a Twirp endpoint that returns errors:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    // Validation errors
    if req.UserId == "" {
        return nil, twirp.InvalidArgument.Error("user_id is required")
    }
    if !isAlphanumeric(req.UserId) {
        return nil, twirp.InvalidArgument.Error("user_id must be alphanumeric")
    }
    if !isAuthorized(ctx, req.UserId) {
        return nil, twirp.PermissionDenied.Error("not allowed to access user profiles")
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

If the endpoint returns a vanilla (non-twirp) error, it will be automatically wrapped using [twirp.InternalErrorWith(err)](https://pkg.go.dev/github.com/twitchtv/twirp#InternalErrorWith).

The following examples are equivalent (the client receives the same internal error).

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    return nil, errors.New("vanilla")
}
```

Is equivalent to wrap the error with the helper:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    return nil, twirp.InternalErrorWith(errors.New("vanilla"))
}
```

Which is also equivalent to building the error from scratch this way:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    err := errors.New("vanilla")
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

Twirp clients return errors that can always be cast to the `twirp.Error` interface. Unpack the error type to access the `Code()`, `Msg()` and `Meta(key)` properties. For example:

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

Use the chainable method [WithMeta(key, val)](https://pkg.go.dev/github.com/twitchtv/twirp#Error.WithMeta) to add extra metadata to a Twirp error. For example:

```go
if unavailable {
    return nil, twirp.Unavailable.Error("taking a nap ...").
        WithMeta("retryable", "true").
        WithMeta("retry_after", "15s")
}
```

Twirp serializes the response as JSON with the additional `meta` field:

```json
// HTTP status: 503
{
  "code": "unavailable",
  "msg": "taking a nap ...",
  "meta": {
    "retryable": "true",
    "retry_after": "15s"
  }
}
```

Metadata is available on the client through the [Meta(key)](https://pkg.go.dev/github.com/twitchtv/twirp#Error.Meta) accessor:

```go
if twerr.Code() == twirp.Unavailable {
    if twerr.Meta("retryable") == "true" {
        fmt.Printf("retry after %s", twerr.Meta("retry_after"))
    }
}
```

Error metadata can only have string values. This is to simplify error parsing by client implementations in multiple platforms. If your service requires errors with complex shapes, consider adding client wrappers on top of the auto-generated clients, or include specific business-logic errors on the Protobuf messages (as part of success responses).
