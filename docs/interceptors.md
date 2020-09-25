---
id: "interceptors"
title: "Interceptors"
sidebar_label: "Interceptors"
---

Interceptors are like middleware, wrapping RPC calls with extra functionality.

In most cases, it is better to use [Hooks](hooks.md) for observability at key points
during a request lifecycle. Hooks do not mutate the request and response structs,
which results in less problems when debugging issues.

Example:

```go
// NewInterceptorMakeSmallHats builds an interceptor that modifies
// calls to MakeHat ignoring the request, and instead always making small hats.
func NewInterceptorMakeSmallHats() twirp.Interceptor {
  return func(next twirp.Method) twirp.Method {
    return func(ctx context.Context, req interface{}) (interface{}, error) {
      if twirp.MethodName(ctx) == "MakeHat" {
        return next(ctx, &haberdasher.Size{Inches: 1})
      }
      return next(ctx, req)
    }
  }
}
```

To wrap all client requests with the interceptor:

```go
client := NewHaberdasherProtobufClient(url, &http.Client{},
  twirp.WithClientInterceptors(NewInterceptorMakeSmallHats()))
```

To wrap all service requests with the interceptor:

```go
server := NewHaberdasherServer(svcImpl,
  twirp.WithServerInterceptors(NewInterceptorMakeSmallHats())
```

Check out
[the godoc for Interceptor](http://godoc.org/github.com/twitchtv/twirp#Interceptor)
for more information.
