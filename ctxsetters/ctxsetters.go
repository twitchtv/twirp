// Package ctxsetters is an implementation detail for twirp generated code, used
// by the generated servers to set values in contexts for later access with the
// twirp package's accessors.
//
// Do not use ctxsetters outside of twirp's generated code.
package ctxsetters

import (
	"context"
	"net/http"
	"strconv"

	"github.com/twitchtv/twirp/internal/contextkeys"
)

func WithMethodName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextkeys.MethodNameKey, name)
}

func WithServiceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextkeys.ServiceNameKey, name)
}

func WithPackageName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextkeys.PackageNameKey, name)
}

func WithStatusCode(ctx context.Context, code int) context.Context {
	return context.WithValue(ctx, contextkeys.StatusCodeKey, strconv.Itoa(code))
}

func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, contextkeys.ResponseWriterKey, w)
}
