---
id: "hooks"
title: "Hooks and Interceptors"
sidebar_label: "Hooks and Interceptors"
---

Twirp main responsibility is [routing and serialization](routing.md), but extra functionality can be plugged in through Hooks and Interceptors. This is useful to do things like log requests, record response times, report metrics, authenticate requests, and so on.

There are multiple ways to inject functionality:

 * [Server Hooks](https://pkg.go.dev/github.com/twitchtv/twirp#ServerHooks): Can be used on the generated server constructor. They provide callbacks for before and after the request is handled. The Error hook is called only if an error was returned by the handler. Every hook receives the request `context.Context` and can return a modified `context.Context` if desired.
 * [Client Hooks](https://pkg.go.dev/github.com/twitchtv/twirp#ClientHooks): Can be used on the generated client constructor. They provide callbacks for before and after the request is sent over the network. The Error hook is called only if an error was retuned through the network.
 * [Interceptors](https://pkg.go.dev/github.com/twitchtv/twirp#Interceptor): Can be used to wrap servers and clients. Interceptors are a form of middleware for Twirp requests. Interceptors can mutate the request and responses, which can enable some powerful integrations, but in most cases, it is better to use Hooks for observability at key points during a request. Mutating the request adds complexity to the request lifecycle.

### Examples


Server Hooks:

```go
// NewLoggingServerHooks logs request and errors to stdout in the service
func NewLoggingServerHooks() *twirp.ServerHooks {
    return &twirp.ServerHooks{
        RequestRouted: func(ctx context.Context) (context.Context, error) {
            method, _ := twirp.MethodName(ctx)
            log.Println("Method: " + method)
            return ctx, nil
        },
        Error: func(ctx context.Context, twerr twirp.Error) context.Context {
            log.Println("Error: " + string(twerr.Code()))
            return ctx
        },
        ResponseSent: func(ctx context.Context) {
            log.Println("Response Sent (error or success)")
        },
    }
}
```

Client Hooks:

```go
// NewLoggingClientHooks logs request and errors to stdout in the client
func NewLoggingClientHooks() *twirp.ClientHooks {
    return &twirp.ClientHooks{
        RequestPrepared: func(ctx context.Context, r *http.Request) (context.Context, error) {
            fmt.Printf("Req: %s %s\n", r.Host, r.URL.Path)
            return ctx, nil
        },
        Error: func(ctx context.Context, twerr twirp.Error) {
            log.Println("Error: " + string(twerr.Code()))
            return ctx
        },
        ResponseReceived: func(ctx context.Context) {
            log.Println("Success")
        },
    }
}
```

Interceptor:

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

Instantiate an example [Haberdasher](example.md) server with hooks and interceptors:

```go
server := NewHaberdasherServer(svcImpl,
    twirp.WithServerInterceptors(NewInterceptorMakeSmallHats()),
    twirp.WithServerHooks(NewLoggingServerHooks()))
```

Instantiate an example [Haberdasher](example.md) client with hooks and interceptors:

```go
client := NewHaberdasherProtobufClient(url, &http.Client{},
    twirp.WithClientInterceptors(NewInterceptorMakeSmallHats()),
    twirp.WithClientHooks(NewLoggingClientHooks()))
```
