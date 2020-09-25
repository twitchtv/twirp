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

// Method is a method that matches the generic form of a Twirp-generated RPC method.
//
// This is used for Interceptors.
type Method func(ctx context.Context, request interface{}) (interface{}, error)

// Interceptor is an interceptor that can be installed on a client or server.
//
// Users can use Interceptors to intercept any RPC.
//
//   func LogInterceptor(l *log.Logger) twirp.Interceptor {
//     return func(next twirp.Method) twirp.Method {
//       return func(ctx context.Context, req interface{}) (interface{}, error) {
//         l.Printf("request: %v", request)
//         resp, err := next(ctx, req)
//         if err != nil {
//           l.Printf("error: %v", err)
//           return nil, err
//         }
//         l.Printf("response: %v", resp)
//         return resp, nil
//       }
//     }
//   }
type Interceptor func(Method) Method

// ChainInterceptors chains the Interceptors.
//
// Returns nil if interceptors is empty.
func ChainInterceptors(interceptors ...Interceptor) Interceptor {
	filtered := make([]Interceptor, 0, len(interceptors))
	for _, interceptor := range interceptors {
		if interceptor != nil {
			filtered = append(filtered, interceptor)
		}
	}
	switch n := len(filtered); n {
	case 0:
		return nil
	case 1:
		return filtered[0]
	default:
		first := filtered[0]
		return func(next Method) Method {
			for i := len(filtered) - 1; i > 0; i-- {
				next = filtered[i](next)
			}
			return first(next)
		}
	}
}
