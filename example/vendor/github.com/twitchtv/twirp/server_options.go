// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.
package twirp

import (
	"context"
)

// ServerOption is a functional option for extending a Twirp client.
type ServerOption func(*ServerOptions)

// ServerOptions encapsulate the configurable parameters on a Twirp client.
type ServerOptions struct {
	Interceptors     []Interceptor
	Hooks            *ServerHooks
	pathPrefix       *string
	JSONSkipDefaults bool
}

func (opts *ServerOptions) PathPrefix() string {
	if opts.pathPrefix == nil {
		return "/twirp" // default prefix
	}
	return *opts.pathPrefix
}

// WithServerInterceptors defines the interceptors for a Twirp server.
func WithServerInterceptors(interceptors ...Interceptor) ServerOption {
	return func(o *ServerOptions) {
		o.Interceptors = append(o.Interceptors, interceptors...)
	}
}

// WithServerHooks defines the hooks for a Twirp server.
func WithServerHooks(hooks *ServerHooks) ServerOption {
	return func(o *ServerOptions) {
		o.Hooks = hooks
	}
}

// ServerHooks is a container for callbacks that can instrument a
// Twirp-generated server. These callbacks all accept a context and return a
// context. They can use this to add to the request context as it threads
// through the system, appending values or deadlines to it.
//
// The RequestReceived and RequestRouted hooks are special: they can return
// errors. If they return a non-nil error, handling for that request will be
// stopped at that point. The Error hook will be triggered, and the error will
// be sent to the client. This can be used for stuff like auth checks before
// deserializing a request.
//
// The RequestReceived hook is always called first, and it is called for every
// request that the Twirp server handles. The last hook to be called in a
// request's lifecycle is always ResponseSent, even in the case of an error.
//
// Details on the timing of each hook are documented as comments on the fields
// of the ServerHooks type.
type ServerHooks struct {
	// RequestReceived is called as soon as a request enters the Twirp
	// server at the earliest available moment.
	RequestReceived func(context.Context) (context.Context, error)

	// RequestRouted is called when a request has been routed to a
	// particular method of the Twirp server.
	RequestRouted func(context.Context) (context.Context, error)

	// ResponsePrepared is called when a request has been handled and a
	// response is ready to be sent to the client.
	ResponsePrepared func(context.Context) context.Context

	// ResponseSent is called when all bytes of a response (including an error
	// response) have been written. Because the ResponseSent hook is terminal, it
	// does not return a context.
	ResponseSent func(context.Context)

	// Error hook is called when an error occurs while handling a request. The
	// Error is passed as argument to the hook.
	Error func(context.Context, Error) context.Context
}

// ChainHooks creates a new *ServerHooks which chains the callbacks in
// each of the constituent hooks passed in. Each hook function will be
// called in the order of the ServerHooks values passed in.
//
// For the erroring hooks, RequestReceived and RequestRouted, any returned
// errors prevent processing by later hooks.
func ChainHooks(hooks ...*ServerHooks) *ServerHooks {
	if len(hooks) == 0 {
		return nil
	}
	if len(hooks) == 1 {
		return hooks[0]
	}
	return &ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			var err error
			for _, h := range hooks {
				if h != nil && h.RequestReceived != nil {
					ctx, err = h.RequestReceived(ctx)
					if err != nil {
						return ctx, err
					}
				}
			}
			return ctx, nil
		},
		RequestRouted: func(ctx context.Context) (context.Context, error) {
			var err error
			for _, h := range hooks {
				if h != nil && h.RequestRouted != nil {
					ctx, err = h.RequestRouted(ctx)
					if err != nil {
						return ctx, err
					}
				}
			}
			return ctx, nil
		},
		ResponsePrepared: func(ctx context.Context) context.Context {
			for _, h := range hooks {
				if h != nil && h.ResponsePrepared != nil {
					ctx = h.ResponsePrepared(ctx)
				}
			}
			return ctx
		},
		ResponseSent: func(ctx context.Context) {
			for _, h := range hooks {
				if h != nil && h.ResponseSent != nil {
					h.ResponseSent(ctx)
				}
			}
		},
		Error: func(ctx context.Context, twerr Error) context.Context {
			for _, h := range hooks {
				if h != nil && h.Error != nil {
					ctx = h.Error(ctx, twerr)
				}
			}
			return ctx
		},
	}
}

// WithServerPathPrefix specifies a different prefix for routing.
// If not specified, the "/twirp" prefix is used by default.
// An empty value "" can be speficied to use no prefix.
// The clients must be configured to send requests using the same prefix.
// URL format: "<baseURL>[<prefix>]/<package>.<Service>/<Method>"
// More info on Twirp docs: https://twitchtv.github.io/twirp/docs/routing.html
func WithServerPathPrefix(prefix string) ServerOption {
	return func(o *ServerOptions) {
		o.pathPrefix = &prefix
	}
}

// WithServerJSONSkipDefaults configures JSON serialization to skip
// unpopulated or default values in JSON responses, which results in
// smaller response sizes. This was the default before v7 and can be
// enabled for full backwards compatibility if required.
// This is now disabled by default, because JSON serialization is
// commonly used for manual debugging, in which case it is useful
// to see the full shape of the response.
func WithServerJSONSkipDefaults(skipDefaults bool) ServerOption {
	return func(o *ServerOptions) {
		o.JSONSkipDefaults = skipDefaults
	}
}
