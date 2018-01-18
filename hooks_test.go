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
	"reflect"
	"testing"
)

func TestChainHooks(t *testing.T) {
	var (
		hook1 = new(ServerHooks)
		hook2 = new(ServerHooks)
		hook3 = new(ServerHooks)
	)

	const key = "key"

	hook1.RequestReceived = func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, key, []string{"hook1"}), nil
	}
	hook2.RequestReceived = func(ctx context.Context) (context.Context, error) {
		v := ctx.Value(key).([]string)
		return context.WithValue(ctx, key, append(v, "hook2")), nil
	}
	hook3.RequestReceived = func(ctx context.Context) (context.Context, error) {
		v := ctx.Value(key).([]string)
		return context.WithValue(ctx, key, append(v, "hook3")), nil
	}

	hook1.RequestRouted = func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, key, []string{"hook1"}), nil
	}

	hook2.ResponsePrepared = func(ctx context.Context) context.Context {
		return context.WithValue(ctx, key, []string{"hook2"})
	}

	chain := ChainHooks(hook1, hook2, hook3)

	ctx := context.Background()

	// When all three chained hooks have a handler, all should be called in order.
	want := []string{"hook1", "hook2", "hook3"}
	haveCtx, err := chain.RequestReceived(ctx)
	if err != nil {
		t.Fatalf("RequestReceived chain has unexpected err %v", err)
	}
	have := haveCtx.Value(key)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("RequestReceived chain has unexpected ctx, have=%v, want=%v", have, want)
	}

	// When only the first chained hook has a handler, it should be called, and
	// there should be no panic.
	want = []string{"hook1"}
	haveCtx, err = chain.RequestRouted(ctx)
	if err != nil {
		t.Fatalf("RequestRouted chain has unexpected err %v", err)
	}
	have = haveCtx.Value(key)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("RequestRouted chain has unexpected ctx, have=%v, want=%v", have, want)
	}

	// When only the second chained hook has a handler, it should be called, and
	// there should be no panic.
	want = []string{"hook2"}
	have = chain.ResponsePrepared(ctx).Value(key)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("RequestRouted chain has unexpected ctx, have=%v, want=%v", have, want)
	}

	// When none of the chained hooks has a handler there should be no panic.
	chain.ResponseSent(ctx)
}
