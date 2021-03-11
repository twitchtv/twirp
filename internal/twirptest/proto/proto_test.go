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

package proto

import (
	context "context"
	fmt "fmt"
	ioutil "io/ioutil"
	http "net/http"
	"net/http/httptest"
	strings "strings"
	"testing"
	"time"

	twirp "github.com/twitchtv/twirp"
)

func TestCompilation(t *testing.T) {
	// Test passes if this package compiles
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("test error")
}

func TestContextCancelError(t *testing.T) {
	// Test passes if this package compiles
	var s Svc
	server := NewSvcServer(s, nil)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Make a request, that will endpoint method
	req, _ := http.NewRequest(http.MethodPost, "http://testing:8080/twirp/twirp.internal.twirptest.proto.Svc/Send", errReader(0))
	req.Header.Set("Accept", "application/protobuf")
	req.Header.Set("Content-Type", "application/protobuf")
	// Associate the cancellable context we just created to the request
	req = req.WithContext(ctx)
	// cancel context
	cancel()
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	expectedErrCode := twirp.ServerHTTPStatusFromErrorCode(twirp.Canceled)
	if w.Code != expectedErrCode {
		t.Errorf("twirp ErrorCode expected to be %q, but found %q", expectedErrCode, w.Code)
	}

	respBytes, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("Could not even read bytes from response: %q", err.Error())
	}

	expectedErrMessage := "context canceled"
	if !strings.Contains(string(respBytes), expectedErrMessage) {
		t.Errorf("twirp client err has unexpected message %q, want %q", string(respBytes), expectedErrMessage)
	}
}

func TestDeadlineExceededError(t *testing.T) {
	// Test passes if this package compiles
	var s Svc
	server := NewSvcServer(s, nil)

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 1*time.Millisecond)

	// Make a request, that will endpoint method
	req, _ := http.NewRequest(http.MethodPost, "http://testing:8080/twirp/twirp.internal.twirptest.proto.Svc/Send", errReader(0))
	req.Header.Set("Accept", "application/protobuf")
	req.Header.Set("Content-Type", "application/protobuf")
	// Associate the cancellable context we just created to the request
	req = req.WithContext(ctx)
	// cancel context
	time.Sleep(2 * time.Millisecond)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	expectedErrCode := twirp.ServerHTTPStatusFromErrorCode(twirp.DeadlineExceeded)
	if w.Code != expectedErrCode {
		t.Errorf("twirp ErrorCode expected to be %q, but found %q", expectedErrCode, w.Code)
	}

	respBytes, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("Could not even read bytes from response: %q", err.Error())
	}

	expectedErrMessage := "context deadline exceeded"
	if !strings.Contains(string(respBytes), expectedErrMessage) {
		t.Errorf("twirp client err has unexpected message %q, want %q", string(respBytes), expectedErrMessage)
	}
}

func TestReadRequestError(t *testing.T) {
	// Test passes if this package compiles
	var s Svc
	server := NewSvcServer(s, nil)

	ctx := context.Background()

	// Make a request, that will endpoint method
	req, _ := http.NewRequest(http.MethodPost, "http://testing:8080/twirp/twirp.internal.twirptest.proto.Svc/Send", errReader(0))
	req.Header.Set("Accept", "application/protobuf")
	req.Header.Set("Content-Type", "application/protobuf")
	// Associate the cancellable context we just created to the request
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// validate error code
	expectedErrCode := twirp.ServerHTTPStatusFromErrorCode(twirp.Internal)
	if w.Code != expectedErrCode {
		t.Errorf("twirp ErrorCode expected to be %q, but found %q", expectedErrCode, w.Code)
	}

	// validate error message
	respBytes, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("Could not even read bytes from response: %q", err.Error())
	}

	expectedErrMessage := "failed to read request body"
	if !strings.Contains(string(respBytes), expectedErrMessage) {
		t.Errorf("twirp client err has unexpected message %q, want %q", string(respBytes), expectedErrMessage)
	}
}
