package logbodies

import (
	"context"
	"log"

	"github.com/twitchtv/twirp"
)

func LogBodies(l *log.Logger) *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestDeserialized: func(ctx context.Context) (context.Context, error) {
			body, ok := twirp.RequestBody(ctx)
			if !ok {
				return ctx, nil
			}
			service, ok := twirp.ServiceName(ctx)
			if !ok {
				service = "unknown"
			}
			method, ok := twirp.MethodName(ctx)
			if !ok {
				method = "unknown"
			}
			l.Printf("%s/%s request: %s", service, method, body)
			return ctx, nil
		},
		ResponsePrepared: func(ctx context.Context) context.Context {
			body, ok := twirp.ResponseBody(ctx)
			if !ok {
				return ctx
			}
			service, ok := twirp.ServiceName(ctx)
			if !ok {
				service = "unknown"
			}
			method, ok := twirp.MethodName(ctx)
			if !ok {
				method = "unknown"
			}
			l.Printf("%s/%s response: %s", service, method, body)
			return ctx
		},
	}
}
