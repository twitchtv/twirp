---
id: "interceptors"
title: "Interceptors"
sidebar_label: "Interceptors"
---

The client and service constructors can use the options
`twirp.WithClientInterceptors(interceptors ...twirp.Interceptor)`
and `twirp.WithServerInterceptors(interceptors ...twirp.Interceptor)`
to plug in additional functionality:

```go
client := NewHaberdasherProtobufClient(url, &http.Client{}, twirp.WithClientInterceptors(NewLogInterceptor(logger.New(os.Stderr, "", 0))))

server := NewHaberdasherServer(svcImpl, twirp.WithServerInterceptors(NewLogInterceptor(logger.New(os.Stderr, "", 0))))

// NewLogInterceptor logs various parts of a request using a standard Logger.
func NewLogInterceptor(l *log.Logger) twirp.Interceptor {
  return func(next twirp.Method) twirp.Method {
    return func(ctx context.Context, req interface{}) (interface{}, error) {
      l.Printf("request: %v", request)
      resp, err := next(ctx, req)
      if err != nil {
        l.Printf("error: %v", err)
        return nil, err
      }
      l.Printf("response: %v", resp)
      return resp, nil
    }
  }
}
```

Check out
[the godoc for `Interceptor`](http://godoc.org/github.com/twitchtv/twirp#Interceptor)
for more information.
