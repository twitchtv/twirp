package main

import (
	"context"

	"github.com/twitchtv/twirp/clientcompat/internal/clientcompat"
)

type clientCompat struct {
	method func(context.Context, *clientcompat.Req) (*clientcompat.Resp, error)
	noop   func(context.Context, *clientcompat.Empty) (*clientcompat.Empty, error)
}

func (c *clientCompat) Method(ctx context.Context, req *clientcompat.Req) (*clientcompat.Resp, error) {
	return c.method(ctx, req)
}

func (c *clientCompat) NoopMethod(ctx context.Context, e *clientcompat.Empty) (*clientcompat.Empty, error) {
	return c.noop(ctx, e)
}
