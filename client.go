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

// HTTPClient is the interface used by generated clients to send HTTP requests.
// It is fulfilled by *(net/http).Client, which is sufficient for most users.
// Users can provide their own implementation for special retry policies.
//
// HTTPClient implementations should not follow redirects. Redirects are
// automatically disabled if *(net/http).Client is passed to client
// constructors. See the withoutRedirects function in this file for more
// details.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientOption func(*ClientOptions)

type ClientOptions struct {
	Client HTTPClient
	Hooks  *ClientHooks
}

type ClientHooks struct {
	// RequestPrepared is called as soon as a request has been created and before it has been sent
	// to the Twirp server.
	RequestPrepared func(context.Context, *http.Request) (context.Context, error)

	// RequestFinished (alternatively could be named ResponseReceived and take in some response from
	// the server as well) is called after a request has finished sending. Since this is terminal, the context is
	// not returned.
	RequestFinished func(context.Context)

	// Error hook is called whenever an error occurs during the sending of a request. The Error is passed
	// as an argument to the hook.
	Error func(context.Context, Error) context.Context
}

func DefaultClientOptions(client HTTPClient) ClientOptions {
	return ClientOptions{
		Client: client,
		Hooks:  nil,
	}
}

func WithClientHooks(hooks *ClientHooks) ClientOption {
	return func(o *ClientOptions) {
		o.Hooks = hooks
	}
}
