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

// Method is a method.
type Method func(ctx context.Context, request interface{}) (interface{}, error)

// Interceptor is a interceptor.
type Interceptor func(Method) Method

// ChainInterceptors chains the Interceptors.
//
// Returns nil if interceptors is empty.
func ChainInterceptors(interceptors ...Interceptor) Interceptor {
	switch n := len(interceptors); n {
	case 0:
		return nil
	case 1:
		return interceptors[0]
	default:
		first := interceptors[0]
		return func(next Method) Method {
			for i := len(interceptors) - 1; i > 0; i-- {
				next = interceptors[i](next)
			}
			return first(next)
		}
	}

}
