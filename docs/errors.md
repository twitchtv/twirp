---
id: "errors"
title: "Errors"
sidebar_label: "Errors"
---

Twirp errors are JSON responses with `code`, `msg` and (optional) `meta` keys:

```json
{
  "code": "internal",
  "msg": "something went wrong",
}
```

Common error codes are `internal`, `not_found`, `invalid_argument` and `permission_denied`. See [twirp.ErrorCode](https://pkg.go.dev/github.com/twitchtv/twirp#ErrorCode) for the full list of available codes.

The [Errors Spec](spec_v7.md#error-codes) has more details about the protocol and HTTP status mapping.

In Go, Twirp errors satisfy the [twirp.Error](https://pkg.go.dev/github.com/twitchtv/twirp#Error) interface. An easy way to instantiate Twirp errors is using the [twirp.NewError](https://pkg.go.dev/github.com/twitchtv/twirp#NewError) constructor.


### Go Clients

Twirp clients always return errors that can be cast to the `twirp.Error` interface.

```go
resp, err := client.MakeHat(ctx, req)
if err != nil {
    if twerr, ok := err.(twirp.Error); ok {
        // twerr.Code()
        // twerr.Msg()
        // twerr.Meta("foobar")
    }
}
```

Transport-level errors (like connection errors) are returned as internal errors by default. If desired, the original client-side error can be unwrapped:

```go
resp, err := client.MakeHat(ctx, req)
if err != nil {
    if twerr, ok := err.(twirp.Error); ok {
        if twerr.Code() == twirp.Internal {
            if transportErr := errors.Unwrap(twerr); transportErr != nil {
                // transportErr could be something like an HTTP connection error
            }
        }
    }
}
```

### Go Services

Example implementation returning Twirp errors:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    if req.UserId == "" {
        return nil, twirp.NewError(twirp.InvalidArgument, "user_id is required")
    }

    user, err := s.DB.FindByID(ctx, req.UserID)
    if err != nil {
        return nil, twirp.WrapError(twirp.NewError(twirp.Internal, "something went wrong"), err)
    }

    if user == nil {
        return nil, twirp.NewError(twirp.NotFound, "user not found")
    }

    return &pb.FindUserResp{
        Login: user.Login,
        // ...
    }, nil
}
```

Errors that can be matched as `twirp.Error` are sent through the wire and returned with the same code in the client.

Regular non-twirp errors are automatically wrapped as internal errors (using [twirp.InternalErrorWith(err)](https://pkg.go.dev/github.com/twitchtv/twirp#InternalErrorWith)). The original error is accessible in service hooks and middleware (e.g. using `errors.Unwrap`). But the original error is NOT serialized through the network; clients cannot access the original error, and will instead receive a `twirp.Error` with code `twirp.Internal`.

Example returning a non-twirp error:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    return nil, errors.New("this non-twirp error will be serialized as a twirp.Internal error")
}
```

Twirp matches with `errors.As(err, &twerr)` to know if a returned error is a `twirp.Error` or not.

**NOTE**: versions older than `v8.1.0` do a type cast `err.(twirp.Error)` instead of matching with `errors.As(err, &twerr)`. This means that wrapped Twirp errors or custom implementations that respond to `As(interface{}) bool` are returned as internal errors, instead of being returned as the appropriate Twirp error. See release [v8.1.0](https://github.com/twitchtv/twirp/releases/tag/v8.1.0) for more details.


### HTTP Errors from Intermediary Proxies

Twirp Clients may receive HTTP responses with non-200 status
from different sources like proxies or load balancers. For example,
a "503 Service Temporarily Unavailable" body, which cannot be
deserialized into a Twirp error.

In those cases, generated Go clients will return `twirp.Error` with a code
depending on the HTTP status of the invalid response:

| HTTP status code         |  Twirp Error Code
| ------------------------ | ------------------
| 3xx (redirects)          | Internalreturn nil, fmt.Errorf("this non-twirp error will
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

Metadata is available on the client as expected:

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

Twirp services can be [muxed with other HTTP services](mux.md). For consistent responses and error codes _outside_ Twirp servers, such as HTTP middleware, you can call [twirp.WriteError](https://pkg.go.dev/github.com/twitchtv/twirp#WriteError).

```go
twerr := twirp.NewError(twirp.Unauthenticated, "invalid token")
twirp.WriteError(respWriter, twerr)
```

As with returned service errors, the error is expected to satisfy the `twirp.Error` interface, otherwise it is wrapped as a `twirp.InternalError`.
