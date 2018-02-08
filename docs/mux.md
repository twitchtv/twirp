---
id: "mux"
title: "Muxing Twirp services"
sidebar_label: "Muxing Twirp with other HTTP services"
---

If you want run your server next to other `http.Handler`s, you'll need to use a
mux. The generated code includes a path prefix that you can use for routing
Twirp requests correctly. It's an exported string const, always as
`<ServiceName>PathPrefix`, and it is the prefix for all Twirp requests.

For example, you could use it with [`http.ServeMux`](https://golang.org/pkg/net/http/#ServeMux) like this:

```go
server := &haberdasherserver.Server{}
twirpHandler := haberdasher.NewHaberdasherServer(server, nil)

mux := http.NewServeMux()
mux.Handle(haberdasher.HaberdasherPathPrefix, twirpHandler)
mux.Handle("/some/other/path", someOtherHandler)

http.ListenAndServe(":8080", mux)
```

You can also serve your Handler on many third-party muxes which accept
`http.Handler`s. For example, on a `goji.Mux`:

```go
server := &haberdasherserver.Server{} // implements Haberdasher interface
twirpHandler := haberdasher.NewHaberdasherServer(server, nil)

mux := goji.NewMux()
mux.Handle(pat.Post(haberdasher.NewHaberdasherPathPrefix+"*"), twirpHandler)
// mux.Handle other things like health checks ...
http.ListenAndServe("localhost:8000", mux)
```
