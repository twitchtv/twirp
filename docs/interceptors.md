---
id: "interceptors"
title: "Interceptors"
sidebar_label: "Interceptors"
---

The service constructor can use the option `twirp.WithServerInterceptors(interceptors ...twirp.Interceptor)`
to plug in additional functionality:

```go
server := NewHaberdasherServer(svcImpl, twirp.WithInterceptor(NewLogInterceptor(logger.New(os.Stderr, "", 0))))

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
