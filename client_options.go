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
	"net/http"
	"reflect"
)

// ClientOption is a functional option for extending a Twirp client.
type ClientOption func(*ClientOptions)

// WithClientHooks defines the hooks for a Twirp client.
func WithClientHooks(hooks *ClientHooks) ClientOption {
	return func(opts *ClientOptions) {
		opts.Hooks = hooks
	}
}

// WithClientInterceptors defines the interceptors for a Twirp client.
func WithClientInterceptors(interceptors ...Interceptor) ClientOption {
	return func(opts *ClientOptions) {
		opts.Interceptors = append(opts.Interceptors, interceptors...)
	}
}

// WithClientPathPrefix specifies a different prefix to use for routing.
// If not specified, the "/twirp" prefix is used by default.
// The service must be configured to serve on the same prefix.
// An empty value "" can be specified to use no prefix.
// URL format: "<baseURL>[<prefix>]/<package>.<Service>/<Method>"
// More info on Twirp docs: https://twitchtv.github.io/twirp/docs/routing.html
func WithClientPathPrefix(prefix string) ClientOption {
	return func(opts *ClientOptions) {
		opts.setOpt("pathPrefix", prefix)
		opts.pathPrefix = &prefix // for code generated before v8.1.0
	}
}

// WithClientLiteralURLs configures the Twirp client to use the exact values
// as defined in the proto file for Service and Method names,
// fixing the issue https://github.com/twitchtv/twirp/issues/244, which is manifested
// when working with Twirp services implemented other languages (e.g. Python) and the proto file definitions
// are not properly following the [Protobuf Style Guide](https://developers.google.com/protocol-buffers/docs/style#services).
// By default (false), Go clients modify the routes by CamelCasing the values. For example,
// with Service: `haberdasher`, Method: `make_hat`, the URLs generated by Go clients are `Haberdasher/MakeHat`,
// but with this option enabled (true) the client will properly use `haberdasher/make_hat` instead.
func WithClientLiteralURLs(b bool) ClientOption {
	return func(opts *ClientOptions) {
		opts.setOpt("literalURLs", b)
		opts.LiteralURLs = b // for code generated before v8.1.0
	}
}

// ClientHooks is a container for callbacks that can instrument a
// Twirp-generated client. These callbacks all accept a context and some return
// a context. They can use this to add to the context, appending values or
// deadlines to it.
//
// The RequestPrepared hook is special because it can return errors. If it
// returns non-nil error, handling for that request will be stopped at that
// point. The Error hook will then be triggered.
//
// The RequestPrepared hook will always be called first and will be called for
// each outgoing request from the Twirp client. The last hook to be called
// will either be Error or ResponseReceived, so be sure to handle both cases in
// your hooks.
type ClientHooks struct {
	// RequestPrepared is called as soon as a request has been created and before
	// it has been sent to the Twirp server.
	RequestPrepared func(context.Context, *http.Request) (context.Context, error)

	// ResponseReceived is called after a request has finished sending. Since this
	// is terminal, the context is not returned. ResponseReceived will not be
	// called in the case of an error being returned from the request.
	ResponseReceived func(context.Context)

	// Error hook is called whenever an error occurs during the sending of a
	// request. The Error is passed as an argument to the hook.
	Error func(context.Context, Error)
}

// ChainClientHooks creates a new *ClientHooks which chains the callbacks in
// each of the constituent hooks passed in. Each hook function will be
// called in the order of the ClientHooks values passed in.
//
// For the erroring hook, RequestPrepared, any returned
// errors prevent processing by later hooks.
func ChainClientHooks(hooks ...*ClientHooks) *ClientHooks {
	if len(hooks) == 0 {
		return nil
	}
	if len(hooks) == 1 {
		return hooks[0]
	}
	return &ClientHooks{
		RequestPrepared: func(ctx context.Context, req *http.Request) (context.Context, error) {
			for _, h := range hooks {
				if h != nil && h.RequestPrepared != nil {
					var err error
					ctx, err = h.RequestPrepared(ctx, req)
					if err != nil {
						return ctx, err
					}
				}
			}
			return ctx, nil
		},
		ResponseReceived: func(ctx context.Context) {
			for _, h := range hooks {
				if h != nil && h.ResponseReceived != nil {
					h.ResponseReceived(ctx)
				}
			}
		},
		Error: func(ctx context.Context, twerr Error) {
			for _, h := range hooks {
				if h != nil && h.Error != nil {
					h.Error(ctx, twerr)
				}
			}
		},
	}
}

// ClientOptions encapsulate the configurable parameters on a Twirp client.
// This type is meant to be used only by generated code.
type ClientOptions struct {
	// Untyped options map. The methods setOpt and ReadOpt are used to set
	// and read options. The options are untyped so when a new option is added,
	// newly generated code can still work with older versions of the runtime.
	m map[string]interface{}

	Hooks        *ClientHooks
	Interceptors []Interceptor

	// Properties below are only used by code that was
	// generated by older versions of Twirp (before v8.1.0).
	// New options with standard types added in the future
	// don't need new properties, they should use ReadOpt.
	LiteralURLs bool
	pathPrefix  *string
}

// ReadOpt extracts an option to a pointer value,
// returns true if the option exists and was extracted.
// This method is meant to be used by generated code,
// keeping the type dependency outside of the runtime.
//
// Usage example:
//
//     opts.setOpt("fooOpt", 123)
//     var foo int
//     ok := opts.ReadOpt("fooOpt", &int)
//
func (opts *ClientOptions) ReadOpt(key string, out interface{}) bool {
	val, ok := opts.m[key]
	if !ok {
		return false
	}

	rout := reflect.ValueOf(out)
	if rout.Kind() != reflect.Ptr {
		panic("ReadOpt(key, out); out must be a pointer but it was not")
	}
	rout.Elem().Set(reflect.ValueOf(val))
	return true
}

// setOpt adds an option key/value. It is used by ServerOption helpers.
// The value can be extracted with ReadOpt by passing a pointer to the same type.
func (opts *ClientOptions) setOpt(key string, val interface{}) {
	if opts.m == nil {
		opts.m = make(map[string]interface{})
	}
	opts.m[key] = val
}

// PathPrefix() is used only by clients generated before v8.1.0
func (opts *ClientOptions) PathPrefix() string {
	if opts.pathPrefix == nil {
		return "/twirp" // default prefix
	}
	return *opts.pathPrefix
}
