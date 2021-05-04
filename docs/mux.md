---
id: "mux"
title: "Muxing Twirp with other HTTP services"
sidebar_label: "Muxing Twirp services"
---

If you want run your server next to other `http.Handler`s, you'll need to use a
mux. The generated code includes a path prefix that you can use for routing
Twirp requests correctly. It's an exported string const, always as
`<ServiceName>PathPrefix`, and it is the prefix for all Twirp requests.

For example, you could use it with [`http.ServeMux`](https://golang.org/pkg/net/http/#ServeMux) like this:

```go
serverImpl := &haberdasherserver.Server{}
twirpHandler := haberdasher.NewHaberdasherServer(serverImpl)

mux := http.NewServeMux()
mux.Handle(twirpHandler.PathPrefix(), twirpHandler)
mux.Handle("/some/other/path", someOtherHandler)

http.ListenAndServe(":8080", mux)
```

You can also serve your Handler on many third-party muxes which accept
`http.Handler`s. For example, on a `goji.Mux`:

```go
serverImpl := &haberdasherserver.Server{} // implements Haberdasher interface
twirpHandler := haberdasher.NewHaberdasherServer(serverImpl)

mux := goji.NewMux()
mux.Handle(pat.Post(twirpHandler.PathPrefix()+"*"), twirpHandler)
// mux.Handle other things like health checks ...
http.ListenAndServe("localhost:8000", mux)
```

### Using a different path prefix

By default, Twirp routes have a "/twirp" path prefix. See
[Routing](routing.md) for more info.

While the URL format can not be customized, the prefix can be changed to allow mounting the service
in different routes. Use the option `twirp.WithServerPathPrefix`:

```go
serverImpl := &haberdasherserver.Server{}
twirpHandler := haberdasher.NewHaberdasherServer(serverImpl,
    twirp.WithServerPathPrefix("/my/custom/prefix"))

mux := http.NewServeMux()
mux.Handle(twirpHandler.PathPrefix(), twirpHandler)
http.ListenAndServe(":8080", mux)
```

The clients must be initialized with the same prefix to send requests to the right routes. Use the
option `twirp.WithClientPathPrefix`:

```go
client := haberdasher.NewHaberdasherProtoClient(s.URL, http.DefaultClient,
    twirp.WithClientPathPrefix("/my/custom/prefix"))
resp, err := c.MakeHat(ctx, &Size{Inches: 1})
```
