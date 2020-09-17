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
	"testing"
)

func TestChainInterceptors(t *testing.T) {
	if chain := ChainInterceptors(); chain != nil {
		t.Errorf("ChainInterceptors(0) expected to be nil, but was %v", chain)
	}

	interceptor1 := func(next Method) Method {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			response, err := next(ctx, request.(string)+"a")
			return response.(string) + "1", err
		}
	}
	interceptor2 := func(next Method) Method {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			response, err := next(ctx, request.(string)+"b")
			return response.(string) + "2", err
		}
	}
	interceptor3 := func(next Method) Method {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			response, err := next(ctx, request.(string)+"c")
			return response.(string) + "3", err
		}
	}
	method := func(ctx context.Context, request interface{}) (interface{}, error) {
		return request.(string) + "x", nil
	}
	for _, testCase := range []struct {
		interceptors []Interceptor
		want         string
	}{
		{
			interceptors: []Interceptor{interceptor1},
			want:         "ax1",
		},
		{
			interceptors: []Interceptor{interceptor1, interceptor2},
			want:         "abx21",
		},
		{
			interceptors: []Interceptor{interceptor1, interceptor2, interceptor3},
			want:         "abcx321",
		},
	} {
		response, err := ChainInterceptors(testCase.interceptors...)(method)(context.Background(), "")
		if err != nil {
			t.Fatalf("ChainInterceptors(%d) method has unexpected err %v", len(testCase.interceptors), err)
		}
		if response != testCase.want {
			t.Errorf("ChainInterceptors(%d) has unexpected value, have=%v, want=%v", len(testCase.interceptors), response, testCase.want)
		}
	}
}
