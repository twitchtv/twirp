---
id: "headers"
title: "Using custom HTTP Headers"
sidebar_label: "Custom HTTP Headers"

---
Sometimes, you need to send custom HTTP headers.

For Twirp, HTTP headers are a transport implementation detail. You should not
have to worry about them, but maybe your HTTP middleware requires them.

If so, there's nothing the Twirp spec that _forbids_ extra headers, so go ahead.
The rest of this doc is a guide on how to do this.

## Client side

### Send HTTP Headers with client requests

Use `twirp.WithHTTPRequestHeaders` to attach the `http.Header` to a particular
`context.Context`, then use that context in the client request:

```go
// Given a client ...
client := haberdasher.NewHaberdasherProtobufClient(addr, &http.Client{})

// Given some headers ...
header := make(http.Header)
header.Set("Twitch-Authorization", "uDRlDxQYbFVXarBvmTncBoWKcZKqrZTY")
header.Set("Twitch-Client-ID", "FrankerZ")

// Attach the headers to a context
ctx := context.Background()
ctx, err := twirp.WithHTTPRequestHeaders(ctx, header)
if err != nil {
  log.Printf("twirp error setting headers: %s", err)
  return
}

// And use the context in the request. Headers will be included in the request!
resp, err := client.MakeHat(ctx, &haberdasher.Size{Inches: 7})
```

### Read HTTP Headers from responses

Twirp client responses are structs that depend only on the Protobuf response.
HTTP headers can not be used by the Twirp client in any way.

However, remember that the Twirp client is instantiated with an `http.Client`,
which can be configured with any `http.RoundTripper` transport. You could make a
RoundTripper that reads some response headers and does something with them.

## Server side

### Send HTTP Headers on server responses

In your server implementation code, set response headers one by one with the
helper `twirp.SetHTTPResponseHeader`, using the same context provided by the
handler. For example:

```go
func (h *myServer) MyRPC(ctx context.Context, req *pb.Req) (*pb.Resp, error) {

  // Add Cache-Control custom header to HTTP response
  err := twirp.SetHTTPResponseHeader(ctx, "Cache-Control", "public, max-age=60")
  if err != nil {
    return nil, twirp.InternalErrorWith(err)
  }

  return &pb.Resp{}, nil
}
```

### Read HTTP Headers from requests

Twirp server methods are abstracted away from HTTP, therefore they don't have
direct access to HTTP Headers.

However, they receive the `http.Request`'s `context.Context` as parameter that
can be modified by HTTP middleware before being used by the Twirp method.

In more detail, you could do the following:

 * Write some middleware (a `func(http.Handler) http.Handler)` that reads the
   header's value and stores it in the request context.
 * Wrap your Twirp server with the middleware you wrote.
 * Inside your service, pull the header value out through the context.

For example, lets say you want to read the 'User-Agent' HTTP header inside a
twirp server method. You might write this middleware:

```go
func WithUserAgent(base http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        ua := r.Header.Get("User-Agent")
        ctx = context.WithValue(ctx, "user-agent", ua)
        r = r.WithContext(ctx)

        base.ServeHTTP(w, r)
    })
}
```

Then, you could wrap your generated Twirp server with this middleware:

```go
h := haberdasher.NewHaberdasherServer(...)
wrapped := WithUserAgent(h)
http.ListenAndServe(":8080", wrapped)
```

Now, in your application code, you would have access to the header through the
context, so you can do whatever you like with it:

```go
func (h *haberdasherImpl) MakeHat(ctx context.Context, req *pb.MakeHatRequest) (*pb.Hat, error) {
    ua := ctx.Value("user-agent").(string)
    log.Printf("user agent: %v", ua)
}
```
