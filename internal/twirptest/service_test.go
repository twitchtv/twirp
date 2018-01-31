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

package twirptest

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/internal/descriptors"
)

func TestServeJSON(t *testing.T) {
	h := PickyHatmaker(1)
	s := httptest.NewServer(NewHaberdasherServer(h, nil))
	defer s.Close()

	client := NewHaberdasherJSONClient(s.URL, http.DefaultClient)

	hat, err := client.MakeHat(context.Background(), &Size{1})
	if err != nil {
		t.Fatalf("JSON Client err=%q", err)
	}
	if hat.Size != 1 {
		t.Errorf("wrong hat size returned")
	}

	_, err = client.MakeHat(context.Background(), &Size{-1})
	if err == nil {
		t.Errorf("JSON Client expected err, got nil")
	}
}

func TestServerJSONWithUnknownFields(t *testing.T) {
	// Haberdasher server that returns same size it was requested
	h := HaberdasherFunc(func(ctx context.Context, s *Size) (*Hat, error) {
		return &Hat{Size: s.Inches}, nil
	})
	s := httptest.NewServer(NewHaberdasherServer(h, nil))
	defer s.Close()

	// Make JSON request with unknown fields ("size" should default to zero-value)
	reqJSON := `{"unknown_field1":"foo", "EXTRASTUFF":"bar"}`
	url := s.URL + HaberdasherPathPrefix + "MakeHat"
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(reqJSON))
	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			t.Fatalf("Closing body: %q", err.Error())
		}
	}()

	// Make sure that the returned hat is valid and has empty (zero-value) size
	respBytes, err := ioutil.ReadAll(resp.Body) // read manually first in case jsonpb.Unmarshal so it can be printed for debugging
	if err != nil {
		t.Fatalf("Could not even read bytes from response: %q", err.Error())
	}
	hat := new(Hat)
	if err = jsonpb.Unmarshal(bytes.NewReader(respBytes), hat); err != nil {
		t.Fatalf("Could not unmarshall response as Hat: %s", respBytes)
	}
	if hat.Size != 0 {
		t.Errorf("Expected empty size (zero-value), found %q", hat.Size)
	}
}

func TestServeProtobuf(t *testing.T) {
	h := PickyHatmaker(1)
	s := httptest.NewServer(NewHaberdasherServer(h, nil))
	defer s.Close()

	client := NewHaberdasherProtobufClient(s.URL, http.DefaultClient)

	hat, err := client.MakeHat(context.Background(), &Size{1})
	if err != nil {
		t.Fatalf("Protobuf Client err=%q", err)
	}
	if hat.Size != 1 {
		t.Errorf("wrong hat size returned")
	}

	_, err = client.MakeHat(context.Background(), &Size{-1})
	if err == nil {
		t.Errorf("Protobuf Client expected err, got nil")
	}
}

type contentTypeOverriderClient struct {
	contentType string
	base HTTPClient
}

func (c *contentTypeOverriderClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", c.contentType)
	return c.base.Do(req)
}

func TestContentTypes(t *testing.T) {
	h := PickyHatmaker(1)
	s := httptest.NewServer(NewHaberdasherServer(h, nil))
	defer s.Close()

	makeClientWithMimeType := func(mime string) Haberdasher {
		return NewHaberdasherJSONClient(s.URL, &contentTypeOverriderClient{
			contentType: mime,
			base: http.DefaultClient,
		})
	}
	expectNoError := func(t *testing.T, mime string) {
		_, err := makeClientWithMimeType(mime).MakeHat(context.Background(), &Size{1})
		if err != nil {
			t.Fatalf("Client using valid mime type %s err=%q", mime, err)
		}
	}

	validMimeTypes := []string{
		"application/json; charset=UTF-8",
		"application/json",
	}
	for _, mime := range validMimeTypes {
		expectNoError(t, mime)
	}

	invalidMimeTypes := []string{
		"application/jsonp",
	}
	for _, mime := range invalidMimeTypes {
		expectBadRouteError(t, makeClientWithMimeType(mime))
	}
}

func TestDeadline(t *testing.T) {
	timeout := 1 * time.Millisecond
	responseTime := 50 * timeout
	h := SlowHatmaker(responseTime)
	s, client := ServerAndClient(h, nil)
	defer s.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		_, err := client.MakeHat(ctx, &Size{1})
		if err == nil {
			t.Errorf("should have timed out, but got nil err")
		}
		close(done)
	}()

	select {
	case <-done:
		// pass
	case <-time.After(responseTime):
		t.Errorf("should have timed out within %s, but %s has elapsed!", timeout, responseTime)
	}
}

// helper type for the requestRecorder to keep track of which hooks have been
// called.
type hookCall string

const (
	received hookCall = "RequestReceived"
	routed   hookCall = "RequestRouted"
	prepared hookCall = "ResponsePrepared"
	sent     hookCall = "ResponseSent"
	errored  hookCall = "Error"
)

type requestRecorder struct {
	sync.Mutex
	calls []hookCall
}

func (r *requestRecorder) reset() {
	r.Lock()
	r.calls = nil
	r.Unlock()
}

func (r *requestRecorder) assertHookCalls(t *testing.T, want []hookCall) {
	r.Lock()
	defer r.Unlock()

	if len(r.calls) != len(want) {
		t.Error("hook calls are wrong")
		t.Logf("have: %v", r.calls)
		t.Logf("want: %v", want)
		t.FailNow()
	}

	for i, haveCall := range r.calls {
		wantCall := want[i]
		if haveCall != wantCall {
			t.Error("hook calls are wrong")
			t.Logf("have: %v", r.calls)
			t.Logf("want: %v", want)
			t.FailNow()
		}
	}
}

func recorderHooks() (*twirp.ServerHooks, *requestRecorder) {
	recs := &requestRecorder{}

	hooks := &twirp.ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			recs.Lock()
			recs.calls = append(recs.calls, received)
			recs.Unlock()
			return ctx, nil
		},
		RequestRouted: func(ctx context.Context) (context.Context, error) {
			recs.Lock()
			recs.calls = append(recs.calls, routed)
			recs.Unlock()
			return ctx, nil
		},
		ResponsePrepared: func(ctx context.Context) context.Context {
			recs.Lock()
			recs.calls = append(recs.calls, prepared)
			recs.Unlock()
			return ctx
		},
		ResponseSent: func(ctx context.Context) {
			recs.Lock()
			recs.calls = append(recs.calls, sent)
			recs.Unlock()
		},
		Error: func(ctx context.Context, _ twirp.Error) context.Context {
			recs.Lock()
			recs.calls = append(recs.calls, errored)
			recs.Unlock()
			return ctx
		},
	}
	return hooks, recs
}

func TestHooks(t *testing.T) {
	hooks, recorder := recorderHooks()
	h := PickyHatmaker(1)

	s := httptest.NewServer(NewHaberdasherServer(h, hooks))
	defer s.Close()
	client := NewHaberdasherProtobufClient(s.URL, http.DefaultClient)

	t.Run("happy path", func(t *testing.T) {
		recorder.reset()
		_, clientErr := client.MakeHat(context.Background(), &Size{1})
		if clientErr != nil {
			t.Fatalf("client err=%q", clientErr)
		}
		recorder.assertHookCalls(t, []hookCall{
			received, routed, prepared, sent,
		})
	})

	t.Run("application error", func(t *testing.T) {
		recorder.reset()
		_, clientErr := client.MakeHat(context.Background(), &Size{-1})
		if clientErr == nil {
			t.Fatal("client err expected with negative Size parameter, but have nil")
		}
		recorder.assertHookCalls(t, []hookCall{
			received, routed, errored, sent,
		})
	})

	t.Run("bad http method", func(t *testing.T) {
		recorder.reset()
		// Use a client that sends GET requests instead of POST.
		rw := &reqRewriter{
			base: http.DefaultTransport,
			rewrite: func(r *http.Request) *http.Request {
				r.Method = "GET"
				return r
			},
		}
		httpClient := &http.Client{Transport: rw}
		client := NewHaberdasherProtobufClient(s.URL, httpClient)

		_, clientErr := client.MakeHat(context.Background(), &Size{-1})
		if clientErr == nil {
			t.Fatal("client err expected with bad HTTP method, but have nil")
		}
		recorder.assertHookCalls(t, []hookCall{
			received, errored, sent,
		})
	})

	t.Run("bad url", func(t *testing.T) {
		recorder.reset()
		// Use a client that sends requests to the wrong URL
		rw := &reqRewriter{
			base: http.DefaultTransport,
			rewrite: func(r *http.Request) *http.Request {
				r.URL.Path = r.URL.Path + "bogus"
				return r
			},
		}
		httpClient := &http.Client{Transport: rw}
		client := NewHaberdasherProtobufClient(s.URL, httpClient)

		_, clientErr := client.MakeHat(context.Background(), &Size{-1})
		if clientErr == nil {
			t.Fatal("client err expected with bad URL, but have nil")
		}
		recorder.assertHookCalls(t, []hookCall{
			received, errored, sent,
		})
	})

	t.Run("missing headers", func(t *testing.T) {
		recorder.reset()
		// Use a client that sends requests without headers
		rw := &reqRewriter{
			base: http.DefaultTransport,
			rewrite: func(r *http.Request) *http.Request {
				r.Header = make(http.Header)
				return r
			},
		}
		httpClient := &http.Client{Transport: rw}
		client := NewHaberdasherProtobufClient(s.URL, httpClient)

		_, clientErr := client.MakeHat(context.Background(), &Size{-1})
		if clientErr == nil {
			t.Fatal("client err expected with missing headers, but have nil")
		}
		recorder.assertHookCalls(t, []hookCall{
			received, errored, sent,
		})
	})

	t.Run("partial request body", func(t *testing.T) {
		recorder.reset()
		// Use a client that sends 1 byte of the body and then stops
		rw := &reqRewriter{
			base: http.DefaultTransport,
			rewrite: func(r *http.Request) *http.Request {
				r.ContentLength = 1
				r.Body = ioutil.NopCloser(io.LimitReader(r.Body, 1))
				return r
			},
		}
		httpClient := &http.Client{Transport: rw}
		client := NewHaberdasherProtobufClient(s.URL, httpClient)

		_, clientErr := client.MakeHat(context.Background(), &Size{-1})
		if clientErr == nil {
			t.Fatal("client err expected with partial request body, but have nil")
		}
		recorder.assertHookCalls(t, []hookCall{
			received, routed, errored, sent,
		})
	})
}

// Test that a Twirp server will work even if it receives a ServiceHooks with a
// nil function for one of its hooks.
func TestNilHooks(t *testing.T) {
	var testcase = func(hooks *twirp.ServerHooks) func(*testing.T) {
		return func(t *testing.T) {
			h := PickyHatmaker(1)
			s := httptest.NewServer(NewHaberdasherServer(h, hooks))
			defer s.Close()
			c := NewHaberdasherProtobufClient(s.URL, http.DefaultClient)
			_, err := c.MakeHat(context.Background(), &Size{1})
			if err != nil {
				t.Fatalf("client err=%q", err)
			}
		}
	}

	h, _ := recorderHooks()
	h.RequestReceived = nil
	t.Run("nil RequestReceived", testcase(h))

	h, _ = recorderHooks()
	h.RequestRouted = nil
	t.Run("nil RequestRouted", testcase(h))

	h, _ = recorderHooks()
	h.ResponsePrepared = nil
	t.Run("nil ResponsePrepared", testcase(h))

	h, _ = recorderHooks()
	h.ResponseSent = nil
	t.Run("nil ResponseSent", testcase(h))

	h, _ = recorderHooks()
	h.Error = nil
	t.Run("nil Error", testcase(h))
}

func TestErroringHooks(t *testing.T) {
	t.Run("RequestReceived error", func(t *testing.T) {
		// Set up hooks that error on request received. The request should be
		// aborted at that point, the Error hook should be triggered, and the client
		// should see an error.
		hooks := &twirp.ServerHooks{}
		hookErr := twirp.NewError(twirp.Unauthenticated, "error in hook")
		errorHookCalled := false
		hooks.RequestReceived = func(ctx context.Context) (context.Context, error) {
			return ctx, hookErr
		}
		hooks.RequestRouted = func(ctx context.Context) (context.Context, error) {
			t.Errorf("request was routed")
			return ctx, nil
		}
		hooks.Error = func(ctx context.Context, err twirp.Error) context.Context {
			if err != hookErr {
				t.Errorf("Error hook did not receive error from RequestReceived. have=%v, want=%v", err, hookErr)
			}
			errorHookCalled = true
			return ctx
		}

		h := PickyHatmaker(1)
		s := httptest.NewServer(NewHaberdasherServer(h, hooks))
		defer s.Close()
		c := NewHaberdasherProtobufClient(s.URL, http.DefaultClient)
		_, err := c.MakeHat(context.Background(), &Size{1})
		if err == nil {
			t.Fatalf("client err=nil, expected=%v", hookErr)
		}
		twerr, ok := err.(twirp.Error)
		if !ok {
			t.Fatalf("client err type=%T, expected twirp.Error", err)
		}

		if twerr.Code() != twirp.Unauthenticated {
			t.Errorf("client err code=%v, expected=%v", twerr.Code(), twirp.Unauthenticated)
		}

		if !errorHookCalled {
			t.Error("Error hook was not triggered")
		}
	})

	t.Run("RequestRouted error", func(t *testing.T) {
		// Set up hooks that error on request routed. The request should be aborted
		// at that point, the Error hook should be triggered, and the client should
		// see an error.
		hooks := &twirp.ServerHooks{}
		hookErr := twirp.NewError(twirp.Unauthenticated, "error in hook")
		errorHookCalled := false
		hooks.RequestRouted = func(ctx context.Context) (context.Context, error) {
			return ctx, hookErr
		}
		hooks.Error = func(ctx context.Context, err twirp.Error) context.Context {
			if err != hookErr {
				t.Errorf("Error hook did not receive error from RequestRouted. have=%v, want=%v", err, hookErr)
			}
			errorHookCalled = true
			return ctx
		}

		h := PickyHatmaker(1)
		s := httptest.NewServer(NewHaberdasherServer(h, hooks))
		defer s.Close()
		c := NewHaberdasherProtobufClient(s.URL, http.DefaultClient)
		_, err := c.MakeHat(context.Background(), &Size{1})
		if err == nil {
			t.Fatalf("client err=nil, expected=%v", hookErr)
		}
		twerr, ok := err.(twirp.Error)
		if !ok {
			t.Fatalf("client err type=%T, expected twirp.Error", err)
		}

		if twerr.Code() != twirp.Unauthenticated {
			t.Errorf("client err code=%v, expected=%v", twerr.Code(), twirp.Unauthenticated)
		}

		if !errorHookCalled {
			t.Error("Error hook was not triggered")
		}
	})
}

func TestInternalErrorPassing(t *testing.T) {
	e := twirp.InternalError("fatal :(")

	h := ErroringHatmaker(e)
	s, c := ServerAndClient(h, nil)
	defer s.Close()

	_, err := c.MakeHat(context.Background(), &Size{})
	if err == nil {
		t.Fatal("expected err, have nil")
	}

	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error type error, have %T", err)
	}

	if twerr.Code() != twirp.Internal {
		t.Errorf("expected error type to be Internal, buf found %q", twerr.Code())
	}
	if twerr.Meta("retryable") != "" {
		t.Errorf("expected error to be non-retryable, but it is (should not have meta)")
	}
	if twerr.Msg() != "fatal :(" {
		t.Errorf("expected error message to be passed through, but have=%q", twerr.Msg())
	}
}

func TestErrorWithRetryableMeta(t *testing.T) {
	eMsg := "try again!"
	e := twirp.NewError(twirp.Unavailable, eMsg)
	e = e.WithMeta("retryable", "true")

	h := ErroringHatmaker(e)
	s, c := ServerAndClient(h, nil)
	defer s.Close()

	_, err := c.MakeHat(context.Background(), &Size{})
	if err == nil {
		t.Fatal("expected err, have nil")
	}

	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error type error, have %T", err)
	}

	if twerr.Meta("retryable") != "true" {
		t.Errorf("expected error to be retryable, but it isnt")
	}
	if twerr.Msg() != eMsg {
		t.Errorf("expected error Msg to be %q, but found %q", eMsg, twerr.Msg())
	}
}

func TestErrorCodePassing(t *testing.T) {
	e := twirp.NewError(twirp.AlreadyExists, "hand-picked ErrorCode")

	h := ErroringHatmaker(e)
	s, c := ServerAndClient(h, nil)
	defer s.Close()

	_, err := c.MakeHat(context.Background(), &Size{})
	if err == nil {
		t.Fatal("expected err, have nil")
	}

	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error type error, have %T", err)
	}

	if twerr.Code() != twirp.AlreadyExists {
		t.Errorf("expected ErrorCode to be passed through to the client to be %q, but have %q", twirp.AlreadyExists, twerr.Code())
	}
}

// Non twirp errors returned by the servers should become twirp Internal errors.
func TestNonTwirpErrorWrappedAsInternal(t *testing.T) {
	e := errors.New("I am not a twirp error, should become internal")

	h := ErroringHatmaker(e)
	s, c := ServerAndClient(h, nil)
	defer s.Close()

	_, err := c.MakeHat(context.Background(), &Size{})
	if err == nil {
		t.Fatal("expected err, found nil")
	}

	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected error type to be 'twirp.Error', but found %T", err)
	}

	if twerr.Code() != twirp.Internal {
		t.Errorf("expected ErrorCode to be %q, but found %q", twirp.Internal, twerr.Code())
	}

	if twerr.Msg() != e.Error() { // NOTE: that twerr.Error() is not e.Error() because it has a "twirp error Internal: *" prefix
		t.Errorf("expected Msg to be %q, but found %q", e.Error(), twerr.Msg())
	}
}

// Clients should be able to connect over HTTPS
func TestConnectTLS(t *testing.T) {
	h := PickyHatmaker(1)
	s := httptest.NewUnstartedServer(NewHaberdasherServer(h, nil))
	s.TLS = &tls.Config{}
	s.StartTLS()
	defer s.Close()

	if !strings.HasPrefix(s.URL, "https") {
		t.Fatal("test server not serving over HTTPS")
	}

	httpsClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	client := NewHaberdasherJSONClient(s.URL, httpsClient)

	hat, err := client.MakeHat(context.Background(), &Size{1})
	if err != nil {
		t.Fatalf("JSON Client err=%q", err)
	}
	if hat.Size != 1 {
		t.Errorf("wrong hat size returned")
	}

	_, err = client.MakeHat(context.Background(), &Size{-1})
	if err == nil {
		t.Errorf("JSON Client expected err, got nil")
	}
}

// It should be possible to serve twirp alongside non-twirp handlers
func TestMuxingTwirpServer(t *testing.T) {
	// Create a healthcheck endpoint. Record that it got called in a boolean.
	healthcheckCalled := false
	healthcheck := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthcheckCalled = true
		w.WriteHeader(200)
		_, err := w.Write([]byte("Looking good, Louis!"))
		if err != nil {
			t.Errorf("err writing response: %s", err)
		}
	})

	// Create a twirp endpoint.
	twirpHandler := NewHaberdasherServer(PickyHatmaker(1), nil)

	// Serve the healthcheck at /health and the twirp handler at the
	// provided URL prefix.
	mux := http.NewServeMux()
	mux.Handle("/health", healthcheck)
	mux.Handle(HaberdasherPathPrefix, twirpHandler)

	s := httptest.NewServer(mux)
	defer s.Close()

	// Try to do a twirp request. It should get routed just fine.
	client := NewHaberdasherJSONClient(s.URL, http.DefaultClient)

	_, twerr := client.MakeHat(context.Background(), &Size{1})
	if twerr != nil {
		t.Errorf("twirp client err=%q", twerr)
	}

	// Try to request the /health endpoint. It should get routed just
	// fine too.
	resp, err := http.Get(s.URL + "/health")
	if err != nil {
		t.Errorf("health check endpoint err=%q", err)
	} else {
		if resp.StatusCode != 200 {
			body, err := ioutil.ReadAll(resp.Body)
			t.Errorf("got a non-200 response from /health: %d", resp.StatusCode)
			t.Logf("response body: %s", body)
			t.Logf("response read err: %s", err)
		}
		if !healthcheckCalled {
			t.Error("health check endpoint was never called")
		}
	}
}

// Default ContextSource should be RequestContextSource, this means that
// when serving in a mux with middleware, the modified request context should
// be available on the method handler.
func TestMuxingTwirpServerDefaultRequestContext(t *testing.T) {
	handlerCalled := false // to verity that the handler assertions were executed (avoid false positives)

	// Make a handler that can check if the context was modified
	twirpHandler := NewHaberdasherServer(HaberdasherFunc(func(ctx context.Context, s *Size) (*Hat, error) {
		handlerCalled = true // verify it was called
		if ctx.Value("modified by middleware") != "yes" {
			t.Error("expected ctx to be modified by the middleware")
		}
		return &Hat{Size: 999}, nil
	}), nil)

	// Wrap with middleware that modifies the request context
	middlewareWrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), "modified by middleware", "yes"))
		twirpHandler.ServeHTTP(w, r)
	})

	// Serve in a mux
	mux := http.NewServeMux()
	mux.Handle(HaberdasherPathPrefix, middlewareWrapper)
	s := httptest.NewServer(mux)
	defer s.Close()

	// And make a request to run the expectations
	client := NewHaberdasherJSONClient(s.URL, http.DefaultClient)
	_, twerr := client.MakeHat(context.Background(), &Size{1})
	if twerr != nil {
		t.Errorf("twirp client err=%q", twerr)
	}
	if !handlerCalled {
		t.Error("For some reason the twirp request did not make it to the handler")
	}
}

// WriteError should allow middleware to easily respond with a properly formatted error response
func TestWriteErrorFromHTTPMiddleware(t *testing.T) {
	// Make a fake server that returns a Twirp error from the HTTP stack, without using an actual Twirp implementation.
	mux := http.NewServeMux()
	mux.HandleFunc(HaberdasherPathPrefix+"MakeHat", func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, twirp.NewError(twirp.Unauthenticated, "You Shall Not Pass!!!"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// A Twirp client is still able to receive the error
	client := NewHaberdasherJSONClient(server.URL, http.DefaultClient)
	_, err := client.MakeHat(context.Background(), &Size{1})
	if err == nil {
		t.Fatal("an error was expected")
	}
	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error type error, have %T", err)
	}
	if twerr.Code() != twirp.Unauthenticated {
		t.Errorf("twirp ErrorCode expected to be %q, but found %q", twirp.Unauthenticated, twerr.Code())
	}
	if twerr.Msg() != "You Shall Not Pass!!!" {
		t.Errorf("twirp client err has unexpected message %q, want %q", twerr.Msg(), "You Shall Not Pass!!!")
	}
}

// WriteError wraps non-twirp errors as twirp.Internal
func TestWriteErrorFromHTTPMiddlewareInternal(t *testing.T) {
	// Make a fake server that returns an error from the HTTP stack, without using an actual Twirp implementation.
	mux := http.NewServeMux()
	mux.HandleFunc(HaberdasherPathPrefix+"MakeHat", func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, errors.New("should become a twirp.Internal"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// A Twirp client is still able to receive the error as a twirp.Internal
	client := NewHaberdasherJSONClient(server.URL, http.DefaultClient)
	_, err := client.MakeHat(context.Background(), &Size{1})
	if err == nil {
		t.Fatal("an error was expected")
	}
	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error type error, have %T", err)
	}
	if twerr.Code() != twirp.Internal {
		t.Errorf("twirp ErrorCode expected to be %q, but found %q", twirp.Internal, twerr.Code())
	}
	if twerr.Msg() != "should become a twirp.Internal" {
		t.Errorf("twirp client err has unexpected message %q, want %q", twerr.Msg(), "should become a twirp.Internal")
	}
}

// If an application panics in its handler, it should return a non-retryable error.
func TestPanickyApplication(t *testing.T) {
	hooks, recorder := recorderHooks()
	s := NewHaberdasherServer(PanickyHatmaker("OH NO!"), hooks)

	// Wrap the twirp server with a handler to stop the panicking from
	// crashing the httptest server and failing our test.
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("http server never panicked")
			}
		}()
		s.ServeHTTP(w, r)
	})

	server := httptest.NewServer(h)
	defer server.Close()

	client := NewHaberdasherJSONClient(server.URL, http.DefaultClient)

	hat, err := client.MakeHat(context.Background(), &Size{1})
	if err == nil {
		t.Logf("hat: %+v", hat)
		t.Fatal("twirp client err is nil for panicking handler")
	}

	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error type error, have %T", err)
	}

	if twerr.Code() != twirp.Internal {
		t.Errorf("twirp ErrorCode expected to be %q, but found %q", twirp.Internal, twerr.Code())
	}
	if twerr.Msg() != "Internal service panic" {
		t.Errorf("twirp client err has unexpected message %q, want %q", twerr.Msg(), "Internal service panic")
	}

	recorder.assertHookCalls(t, []hookCall{received, routed, errored, sent})
}

func TestCustomRequestHeaders(t *testing.T) {
	// Create a set of headers to be sent on all requests
	customHeader := make(http.Header)
	customHeader.Set("key1", "val1")
	customHeader.Add("multikey", "val1")
	customHeader.Add("multikey", "val2")

	haberdasher := NewHaberdasherServer(PickyHatmaker(1), nil)
	// Make a wrapping handler that checks headers for this test
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All key-vals in the custom header should appear in the request
		for k, v := range customHeader {
			if !reflect.DeepEqual(r.Header[k], v) {
				t.Errorf("missing header  key=%q  wanted-val=%q  have-val=%q", k, v, r.Header[k])
			}
		}
		haberdasher.ServeHTTP(w, r)
	})

	s := httptest.NewServer(handler)
	defer s.Close()

	clients := map[string]Haberdasher{
		"protobuf": NewHaberdasherProtobufClient(s.URL, http.DefaultClient),
		"json":     NewHaberdasherJSONClient(s.URL, http.DefaultClient),
	}
	for name, c := range clients {
		t.Logf("client=%q", name)
		ctx := context.Background()
		ctx, err := twirp.WithHTTPRequestHeaders(ctx, customHeader)
		if err != nil {
			t.Fatalf("%q client WithHTTPRequestHeaders err=%q", name, err)
		}
		_, err = c.MakeHat(ctx, &Size{1})
		if err != nil {
			t.Errorf("%q client err=%q", name, err)
		}
	}
}

func TestCustomResponseHeaders(t *testing.T) {
	// service that adds headers key1 and key2
	haberdasher := NewHaberdasherServer(HaberdasherFunc(func(ctx context.Context, s *Size) (*Hat, error) {
		var err error
		errMsg := "unexpected error returned by SetHTTPResponseHeader: "

		err = twirp.SetHTTPResponseHeader(ctx, "key1", "val1")
		if err != nil {
			t.Fatalf(errMsg + err.Error())
		}

		err = twirp.SetHTTPResponseHeader(ctx, "key2", "before_override")
		if err != nil {
			t.Fatalf(errMsg + err.Error())
		}
		err = twirp.SetHTTPResponseHeader(ctx, "key2", "val2") // should override
		if err != nil {
			t.Fatalf(errMsg + err.Error())
		}

		childContext := context.WithValue(ctx, "foo", "var")
		err = twirp.SetHTTPResponseHeader(childContext, "key3", "val3")
		if err != nil {
			t.Fatalf(errMsg + err.Error())
		}

		err = twirp.SetHTTPResponseHeader(context.Background(), "key4", "should_be_ignored")
		if err != nil {
			t.Fatalf(errMsg + err.Error())
		}

		err = twirp.SetHTTPResponseHeader(context.Background(), "Content-Type", "should_return_error")
		if err == nil {
			t.Fatalf("SetHTTPResponseHeader expected to return an error on Content-Type header, found nil")
		}

		return &Hat{Size: 999}, nil
	}), nil)

	// Make a wrapping handler that checks header responses for this test
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		haberdasher.ServeHTTP(w, r)

		if w.Header().Get("key1") != "val1" {
			t.Errorf("expected 'key1' header to be 'val1', but found %q", w.Header().Get("key1"))
		}
		if w.Header().Get("key2") != "val2" {
			t.Errorf("expected 'key2' header to be 'val2', but found %q", w.Header().Get("key2"))
		}
		if w.Header().Get("key3") != "val3" {
			t.Errorf("expected 'key3' header to be 'val3', but found %q", w.Header().Get("key3"))
		}
		if w.Header().Get("key4") == "should_be_ignored" {
			t.Error("expected 'key4' header to be empty, it should be ignored if the context is not coming from the handler")
		}
	})

	s := httptest.NewServer(handler)
	defer s.Close()

	c := NewHaberdasherProtobufClient(s.URL, http.DefaultClient)
	resp, err := c.MakeHat(context.Background(), &Size{1})
	if err != nil {
		t.Errorf("unexpected service error: %q", err)
	}
	if resp.Size != 999 { // make sure that the fake handler and its assertions were called
		t.Errorf("expected resp.Size to be 999 (coming from fake handler), but found %d", resp.Size)
	}
}

// A nil response should cause an 'Internal Error' response, not a
// panic.
func TestNilResponse(t *testing.T) {
	h := NewHaberdasherServer(NilHatmaker(), nil)

	panicChecker := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				t.Errorf("handler panicked: %s", p)
			}
		}()
		h.ServeHTTP(w, r)
	})
	s := httptest.NewServer(panicChecker)
	defer s.Close()

	clients := map[string]Haberdasher{
		"protobuf": NewHaberdasherProtobufClient(s.URL, http.DefaultClient),
		"json":     NewHaberdasherJSONClient(s.URL, http.DefaultClient),
	}
	for name, c := range clients {
		_, err := c.MakeHat(context.Background(), &Size{1})
		if err == nil {
			t.Errorf("%q client err=nil, which is unexpected", name)
		}
	}
}


var expectBadRouteError = func(t *testing.T, client Haberdasher) {
	_, err := client.MakeHat(context.Background(), &Size{1})
	if err == nil {
		t.Fatalf("err=nil, expected bad_route")
	}

	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("err has type=%T, expected twirp.Error", err)
	}

	if twerr.Code() != twirp.BadRoute {
		t.Errorf("err has code=%v, expected %v", twerr.Code(), twirp.BadRoute)
	}
}

func TestBadRoute(t *testing.T) {
	h := PickyHatmaker(1)
	s := httptest.NewServer(NewHaberdasherServer(h, nil))
	defer s.Close()

	// Create a transport that we can use to force the wrong HTTP method and URL
	// in test cases.
	rw := &reqRewriter{base: http.DefaultTransport}
	httpClient := &http.Client{Transport: rw}

	clients := map[string]Haberdasher{
		"json":     NewHaberdasherJSONClient(s.URL, httpClient),
		"protobuf": NewHaberdasherProtobufClient(s.URL, httpClient),
	}

	for name, client := range clients {
		t.Run(name+" client", func(t *testing.T) {
			t.Run("good route", func(t *testing.T) {
				rw.rewrite = func(r *http.Request) *http.Request { return r }
				_, err := client.MakeHat(context.Background(), &Size{1})
				if err != nil {
					t.Errorf("unexpected error with vanilla client and transport: %v", err)
				}
			})

			t.Run("bad URL path", func(t *testing.T) {
				rw.rewrite = func(r *http.Request) *http.Request {
					r.URL.Path = r.URL.Path + "bogus"
					return r
				}
				expectBadRouteError(t, client)
			})

			t.Run("bad HTTP method", func(t *testing.T) {
				rw.rewrite = func(r *http.Request) *http.Request {
					r.Method = "GET"
					return r
				}
				expectBadRouteError(t, client)
			})

			t.Run("bad content-type", func(t *testing.T) {
				rw.rewrite = func(r *http.Request) *http.Request {
					r.Header.Set("Content-Type", "application/bogus")
					return r
				}
				expectBadRouteError(t, client)
			})
		})
	}
}

// reqRewriter is a http.RoundTripper which can be used to mess with a request
// before it actually gets sent. This is useful as a transport for http.Clients
// in tests because it lets us modify the HTTP method, path, and headers of a
// request, while still being sure that the other unchanged fields are set by a
// generated client.
type reqRewriter struct {
	base    http.RoundTripper
	rewrite func(r *http.Request) *http.Request
}

func (rw *reqRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	req = rw.rewrite(req)
	return rw.base.RoundTrip(req)
}

func TestReflection(t *testing.T) {
	h := PickyHatmaker(1)
	s := NewHaberdasherServer(h, nil)

	t.Run("ServiceDescriptor", func(t *testing.T) {
		fd, sd, err := descriptors.ServiceDescriptor(s)
		if err != nil {
			t.Fatalf("unable to load descriptor: %v", err)
		}
		if have, want := fd.GetPackage(), "twirp.internal.twirptest"; have != want {
			t.Errorf("bad package name, have=%q, want=%q", have, want)
		}

		if have, want := sd.GetName(), "Haberdasher"; have != want {
			t.Errorf("bad service name, have=%q, want=%q", have, want)
		}
		if len(sd.Method) != 1 {
			t.Errorf("unexpected number of methods, have=%d, want=1", len(sd.Method))
		}
		if have, want := sd.Method[0].GetName(), "MakeHat"; have != want {
			t.Errorf("bad method name, have=%q, want=%q", have, want)
		}
	})
	t.Run("ProtoGenTwirpVersion", func(t *testing.T) {
		// Should match whatever is in the file at protoc-gen-twirp/version.go
		file, err := ioutil.ReadFile("../gen/version.go")
		if err != nil {
			t.Fatalf("unable to load version file: %v", err)
		}
		rexp, err := regexp.Compile(`const Version = "(.*)"`)
		if err != nil {
			t.Fatalf("unable to compile twirpVersion regex: %v", err)
		}
		matches := rexp.FindSubmatch(file)
		if matches == nil {
			t.Fatal("unable to find twirp version from version.go file")
		}

		want := string(matches[1])
		have := s.ProtocGenTwirpVersion()
		if have != want {
			t.Errorf("bad version, have=%q, want=%q", have, want)
		}
	})
}

func TestContextValues(t *testing.T) {
	h := HaberdasherFunc(func(ctx context.Context, _ *Size) (*Hat, error) {
		const (
			wantPkg    = "twirp.internal.twirptest"
			wantSvc    = "Haberdasher"
			wantMethod = "MakeHat"
		)
		if pkg, _ := twirp.PackageName(ctx); pkg != wantPkg {
			t.Errorf("twirp.PackageName(ctx)  have=%q, want=%q", pkg, wantPkg)
		}
		if svc, _ := twirp.ServiceName(ctx); svc != wantSvc {
			t.Errorf("twirp.ServiceName(ctx)  have=%q, want=%q", svc, wantSvc)
		}
		if meth, _ := twirp.MethodName(ctx); meth != wantMethod {
			t.Errorf("twirp.MethodName(ctx)  have=%q, want=%q", meth, wantMethod)
		}
		return &Hat{}, nil
	})
	s := httptest.NewServer(NewHaberdasherServer(h, nil))
	defer s.Close()

	client := NewHaberdasherProtobufClient(s.URL, http.DefaultClient)

	_, err := client.MakeHat(context.Background(), &Size{1})
	if err != nil {
		t.Errorf("Client err=%q", err)
	}
}
