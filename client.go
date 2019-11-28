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
)

type ClientOption func(*ClientOptions)

type ClientOptions struct {
	Hooks *ClientHooks
}

type ClientHooks struct {
	// RequestPrepared is called as soon as a request has been created and before it has been sent
	// to the Twirp server.
	RequestPrepared func(context.Context, *http.Request) (context.Context, error)

	// ResponseReceived is called after a request has finished sending. Since this is terminal, the context is
	// not returned.
	ResponseReceived func(context.Context)

	// Error hook is called whenever an error occurs during the sending of a request. The Error is passed
	// as an argument to the hook.
	Error func(context.Context, Error)
}

func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		Hooks: nil,
	}
}

func WithClientHooks(hooks *ClientHooks) ClientOption {
	return func(o *ClientOptions) {
		o.Hooks = hooks
	}
}
