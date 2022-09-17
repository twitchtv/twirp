---
id: "errors"
title: "Errors"
sidebar_label: "Errors"
---

Twirp errors are JSON responses with `code`, `msg` and (optional) `meta` keys.

Example of `internal` error:

```json
{
  "code": "internal",
  "msg": "something went wrong",
}
```

Example of `not_found` error with metadata:

```json
{
  "code": "not_found",
  "msg": "user not found",
  "meta": {
    "user_id": "123",
    "retry": "no",
  },
}
```

Valid Twirp Error Codes:

 * `internal` (500)
 * `not_found` (404)
 * `invalid_argument` (400)
 * `unauthenticated` (401)
 * `permission_denied` (403)
 * `already_exists` (409)
 * ... see all codes on the [Errors Spec](spec_v7.md#error-codes).

Twirp services map each error code to a equivalent HTTP status to make it easy to check for errors on middleware.

In Go, Twirp errors satisfy the [twirp.Error](https://pkg.go.dev/github.com/twitchtv/twirp#Error) interface. The `twirp` package provides error constructors from the codes (e.g. `twirp.Internal.Error("something went wrong")`) and a generic constructor [twirp.NewError](https://pkg.go.dev/github.com/twitchtv/twirp#NewError). Check the [errors.go](https://github.com/twitchtv/twirp/blob/main/errors.go) file for more examples!


### Returning Twirp errors from Go Services

Example service implementation returning Twirp errors:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    if req.UserId == "" {
        return nil, twirp.InvalidArgument.Error("user_id is required").WithMeta("arg", "user_id")
    }

    user, err := s.DB.FindByID(ctx, req.UserID)
    if errors.Is(err, DB_NOT_FOUND) {
        return nil, twirp.NotFound.Error("user not found")
    }
    if err != nil {
        return nil, twirp.Internal.Errorf("DB error: %w", err)
    }

    return &pb.FindUserResp{
        Login: user.Login,
        // ...
    }, nil
}
```

Error values that implement the [twirp.Error](https://pkg.go.dev/github.com/twitchtv/twirp#Error) interface are sent through the wire and returned with the same `code`, `msg` and `meta` in the client.

Regular non-twirp errors are automatically wrapped as internal errors (using [twirp.InternalErrorWith(err)](https://pkg.go.dev/github.com/twitchtv/twirp#InternalErrorWith)). The original error can be unwrapped and accessed on service hooks and middleware (e.g. using `errors.Unwrap`). But the original error is NOT serialized through the network; clients cannot access the original error, and will instead receive a `twirp.Error` with code `internal`.

Example returning a non-twirp error:

```go
func (s *Server) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResp, error) {
    return nil, errors.New("will be serialized as twirp.Internal")
}
```

### Handling errors on Go Clients

Twirp clients return errors that can always be cast to the `twirp.Error` interface. Unpack the error type to access the `Code()`, `Msg()` and `Meta(key)` properties:

```go
resp, err := client.FindUser(ctx, req)
if err != nil {
    if twerr, ok := err.(twirp.Error); ok {
        if twerr.Code() == twirp.NotFound {
            fmt.Println("not found")
        } else {
            fmt.Printf("Twirp error %s: %q", twerr.Code(), twerr.Msg())
        }
    }
}
```

You can also use [errors.Is](https://pkg.go.dev/errors#Is) and [errors.As](https://pkg.go.dev/errors#As) to check and unwrap Twirp errors:

```go
resp, err := client.MakeHat(ctx, req)
var twerr twirp.Error
if errors.As(err, &twerr) {
    if twerr.Code() == twirp.NotFound {
        fmt.Println("not found")
    } else {
        fmt.Printf("Twirp error %s: %q", twerr.Code(), twerr.Msg())
    }
}
```

Transport-level errors (like connection errors) are returned as internal errors by default. If desired, the original client-side network error can be unwrapped:

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

If the server error was wrapping an original service error, that can not be accessed in the client; only the twirp error properties `code`, `msg` and `meta` are serialized over the network.


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
    return nil, twirp.Unavailable.Error("taking a nap ...").
        WithMeta("retryable", "true").
        WithMeta("retry_after", "15s")
}
```

Metadata is available on the client as expected:

```go
if twerr.Code() == twirp.Unavailable {
    if twerr.Meta("retryable") == "true" {
        fmt.Printf("retry after %s", twerr.Meta("retry_after"))
    }
}
```

Error metadata can only have string values. This is to simplify error parsing by clients. If your service requires errors with complex metadata, you should consider adding client wrappers on top of the auto-generated clients, or add specific business-logic errors as a speficic error field on the Protobuf messages.


### Writing HTTP Errors outside Twirp services

Twirp services can be [muxed with other HTTP services](mux.md). For consistent responses and error codes _outside_ Twirp servers, such as HTTP middleware, you can call [twirp.WriteError](https://pkg.go.dev/github.com/twitchtv/twirp#WriteError).

```go
twerr := twirp.Unauthenticated.Error("invalid token")
twirp.WriteError(respWriter, twerr)
```

As with returned service errors, the error is expected to satisfy the `twirp.Error` interface, otherwise it is wrapped as a `twirp.InternalError`.
