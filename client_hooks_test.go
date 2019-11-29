package twirp

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestChainClientHooks(t *testing.T) {
	var (
		hook1 = new(ClientHooks)
		hook2 = new(ClientHooks)
		hook3 = new(ClientHooks)

		responseReceivedCalled []string
		errorCalled            []string
	)

	const key = "key"

	hook1.RequestPrepared = func(ctx context.Context, req *http.Request) (context.Context, error) {
		return context.WithValue(ctx, key, []string{"hook1"}), nil
	}
	hook2.RequestPrepared = func(ctx context.Context, req *http.Request) (context.Context, error) {
		v := ctx.Value(key).([]string)
		return context.WithValue(ctx, key, append(v, "hook2")), nil
	}
	hook3.RequestPrepared = func(ctx context.Context, req *http.Request) (context.Context, error) {
		v := ctx.Value(key).([]string)
		return context.WithValue(ctx, key, append(v, "hook3")), nil
	}

	hook1.ResponseReceived = func(ctx context.Context) {
		responseReceivedCalled = append(responseReceivedCalled, "hook1")
	}

	hook2.Error = func(ctx context.Context, twerr Error) {
		errorCalled = append(errorCalled, "hook2")
	}

	chain := ChainClientHooks(hook1, hook2, hook3)

	ctx := context.Background()

	// When all three chained hooks have a handler, all should be called in order.
	want := []string{"hook1", "hook2", "hook3"}
	haveCtx, err := chain.RequestPrepared(ctx, new(http.Request))
	if err != nil {
		t.Fatalf("RequestPrepared chain has unexpected err %v", err)
	}
	have := haveCtx.Value(key)
	if !reflect.DeepEqual(have, want) {
		t.Errorf("RequestPrepared chain has unexpected ctx, have=%v, want=%v", have, want)
	}

	// When only the first chained hook has a handler, it should be called, and
	// there should be no panic.
	want = []string{"hook1"}
	chain.ResponseReceived(ctx)
	if have := responseReceivedCalled; !reflect.DeepEqual(have, want) {
		t.Errorf("unexpected hooks called, have: %v, want: %v", have, want)
	}

	// When only the second chained hook has a handler, it should be called, and
	// there should be no panic.
	want = []string{"hook2"}
	chain.Error(ctx, InternalError("whoops"))
	if have := errorCalled; !reflect.DeepEqual(have, want) {
		t.Errorf("unexpected hooks called, have: %v, want: %v", have, want)
	}

	// When none of the chained hooks has a handler there should be no panic.
	errorCalled = nil
	hook2.Error = nil
	chain.Error(ctx, InternalError("whoops"))
	if have, want := 0, len(errorCalled); have != want {
		t.Errorf("unexpected number of calls, have: %d, want: %d", have, want)
	}
}
