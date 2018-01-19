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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/clientcompat/internal/clientcompat"
)

var (
	failures  = 0
	successes = 0
)

func main() {
	clientBin := flag.String("client", "", "client binary")
	flag.Parse()

	if *clientBin == "" {
		log.Fatal("-client must be specified")
	}

	cc, s := newServer()
	defer s.Close()

	testNoop(cc, s, *clientBin)
	testMethod(cc, s, *clientBin)
	testInvalidErrorHandling(*clientBin)

	if failures > 0 {
		fmt.Printf("FAILED with %d failures, %d successes\n", failures, successes)
		os.Exit(1)
	}
	fmt.Printf("PASSED with %d failures, %d successes\n", failures, successes)
}

func newServer() (*clientCompat, *httptest.Server) {
	cc := &clientCompat{}
	s := clientcompat.NewCompatServiceServer(cc, nil)
	return cc, httptest.NewServer(s)
}

func startTest(name string) {
	fmt.Printf("Testing %v... ", name)
}

func fail(msg string, args ...interface{}) {
	failures++
	fmt.Printf("FAIL: "+msg+"\n", args...)
}

func pass() {
	successes++
	fmt.Printf("PASS\n")
}

func testNoop(cc *clientCompat, s *httptest.Server, clientBin string) {

	type noop func(context.Context, *clientcompat.Empty) (*clientcompat.Empty, error)

	testcase := func(name string, f noop, wantErrCode string) {
		startTest(name)
		cc.noop = f
		_, haveErrCode, err := runClientNoop(s.URL, clientBin)
		if err != nil {
			fail("error: %v", err)
			return
		}

		switch {
		case wantErrCode == "" && haveErrCode != "":
			fail("client reported twirp error %q when server did not error", haveErrCode)
		case wantErrCode != "" && haveErrCode == "":
			fail("client did not report err code when server errored, expected %q", wantErrCode)
		case wantErrCode != haveErrCode:
			fail("client reported wrong error code %q, want %q", haveErrCode, wantErrCode)
		default:
			pass()
		}
	}

	testcase(
		"noop without error",
		func(context.Context, *clientcompat.Empty) (*clientcompat.Empty, error) {
			return &clientcompat.Empty{}, nil
		},
		"",
	)

	for _, code := range []twirp.ErrorCode{
		twirp.Canceled, twirp.Unknown, twirp.InvalidArgument, twirp.DeadlineExceeded,
		twirp.NotFound, twirp.BadRoute, twirp.AlreadyExists, twirp.PermissionDenied,
		twirp.Unauthenticated, twirp.ResourceExhausted, twirp.FailedPrecondition,
		twirp.Aborted, twirp.OutOfRange, twirp.Unimplemented, twirp.Internal,
		twirp.Unavailable, twirp.DataLoss,
	} {
		testcase(
			fmt.Sprintf("%q error parsing", code),
			func(context.Context, *clientcompat.Empty) (*clientcompat.Empty, error) {
				return nil, twirp.NewError(code, "failed")
			},
			string(code),
		)
	}
}

func testMethod(cc *clientCompat, s *httptest.Server, clientBin string) {

	type method func(context.Context, *clientcompat.Req) (*clientcompat.Resp, error)

	testcase := func(name string, req *clientcompat.Req, f method, wantResp *clientcompat.Resp, wantErrCode string) {
		startTest(name)

		called := false

		cc.method = func(ctx context.Context, req *clientcompat.Req) (*clientcompat.Resp, error) {
			called = true
			return f(ctx, req)
		}

		resp, haveErrCode, err := runClientMethod(s.URL, clientBin, req)
		if err != nil {
			fail("error: %v", err)
			return
		}

		if !called {
			fail("RPC Method was not called on server")
			return
		}

		switch {
		case wantErrCode == "" && haveErrCode != "":
			fail("client reported twirp error %q when server did not error", haveErrCode)
			return
		case wantErrCode != "" && haveErrCode == "":
			fail("client did not report err code when server errored, expected %q", wantErrCode)
			return
		case wantErrCode != haveErrCode:
			fail("client reported wrong error code %q, want %q", haveErrCode, wantErrCode)
			return
		}

		if !reflect.DeepEqual(resp, wantResp) {
			fail("client has wrong response, have %+v want %+v", resp, wantResp)
			return
		}

		pass()
	}

	testcase(
		"empty value",
		&clientcompat.Req{},
		func(context.Context, *clientcompat.Req) (*clientcompat.Resp, error) {
			return &clientcompat.Resp{}, nil
		},
		&clientcompat.Resp{},
		"",
	)

	testcase(
		"request value formatting",
		&clientcompat.Req{
			V: "value",
		},
		func(_ context.Context, req *clientcompat.Req) (*clientcompat.Resp, error) {
			if req.V != "value" {
				return nil, twirp.InvalidArgumentError("V", "should be 'value'")
			}
			return &clientcompat.Resp{V: 1}, nil
		},
		&clientcompat.Resp{V: 1},
		"",
	)
}

func testInvalidErrorHandling(clientBin string) {
	startTest("handling invalid error formatting from server")
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := w.Write([]byte("garbage"))
		if err != nil {
			panic(err)
		}
	}))
	defer s.Close()
	_, errCode, err := runClientNoop(s.URL, clientBin)
	if err != nil {
		fail("err: %v", err)
	} else if errCode != "internal" {
		fail("wrong error code: %v", errCode)
	}
	pass()
}
