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

package twirp_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/twitchtv/twirp"
)

func TestErrorConstructors(t *testing.T) {
	var twerr twirp.Error
	err := errors.New("The OG error")

	// code.Error

	twerr = twirp.NotFound.Error("oops")
	assertTwirpError(t, twerr, twirp.NotFound, "oops")

	// code.Errorf

	twerr = twirp.Aborted.Errorf("oops %d %s", 11, "times")
	assertTwirpError(t, twerr, twirp.Aborted, "oops 11 times")

	twerr = twirp.Internal.Errorf("oops: %w", err)
	assertTwirpError(t, twerr, twirp.Internal, "oops: The OG error")
	if !errors.Is(twerr, err) {
		t.Errorf("expected wrap the original error")
	}

	// twirp.NewError

	twerr = twirp.NewError(twirp.NotFound, "oops")
	assertTwirpError(t, twerr, twirp.NotFound, "oops")

	// twirp.NewErrorf

	twerr = twirp.NewErrorf(twirp.Aborted, "oops %d %s", 11, "times")
	assertTwirpError(t, twerr, twirp.Aborted, "oops 11 times")

	twerr = twirp.NewErrorf(twirp.Internal, "oops: %w", err)
	assertTwirpError(t, twerr, twirp.Internal, "oops: The OG error")
	if !errors.Is(twerr, err) {
		t.Errorf("expected wrap the original error")
	}

	// Convenience constructors

	twerr = twirp.NotFoundError("oops")
	assertTwirpError(t, twerr, twirp.NotFound, "oops")

	twerr = twirp.InvalidArgumentError("my_arg", "is invalid")
	assertTwirpError(t, twerr, twirp.InvalidArgument, "my_arg is invalid")
	assertTwirpErrorMeta(t, twerr, "argument", "my_arg")

	twerr = twirp.RequiredArgumentError("my_arg")
	assertTwirpError(t, twerr, twirp.InvalidArgument, "my_arg is required")
	assertTwirpErrorMeta(t, twerr, "argument", "my_arg")

	twerr = twirp.InternalError("oops")
	assertTwirpError(t, twerr, twirp.Internal, "oops")

	twerr = twirp.InternalErrorf("oops: %w", err)
	assertTwirpError(t, twerr, twirp.Internal, "oops: The OG error")
	if !errors.Is(twerr, err) {
		t.Errorf("expected wrap the original error")
	}

	twerr = twirp.InternalErrorWith(err)
	assertTwirpError(t, twerr, twirp.Internal, "The OG error")
	if !errors.Is(twerr, err) {
		t.Errorf("expected wrap the original error")
	}
	assertTwirpErrorMeta(t, twerr, "cause", "*errors.errorString")
}

func TestWithMetaRaces(t *testing.T) {
	err := twirp.NewError(twirp.Internal, "msg")
	err = err.WithMeta("k1", "v1")

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			_ = err.WithMeta(fmt.Sprintf("k-%d", i), "v")
			wg.Done()
		}(i)
	}

	wg.Wait()

	if len(err.MetaMap()) != 1 {
		t.Errorf("err was mutated")
	}
}

func TestPkgErrorCause(t *testing.T) {
	rootCause := pkgerrors.New("this is only a test")
	twerr := twirp.InternalErrorWith(rootCause)
	cause := pkgerrors.Cause(twerr)
	if cause != rootCause {
		t.Errorf("got wrong cause for err. have=%q, want=%q", cause, rootCause)
	}
}

func TestWrapError(t *testing.T) {
	rootCause := errors.New("cause")
	twerr := twirp.NewError(twirp.NotFound, "it ain't there")
	err := twirp.WrapError(twerr, rootCause)
	cause := pkgerrors.Cause(err)
	if cause != rootCause {
		t.Errorf("got wrong cause. got=%q, want=%q", cause, rootCause)
	}
	wantMsg := "twirp error not_found: it ain't there"
	if gotMsg := err.Error(); gotMsg != wantMsg {
		t.Errorf("got wrong error text. got=%q, want=%q", gotMsg, wantMsg)
	}
}

type myError string

func (e myError) Error() string {
	return string(e)
}

func TestInternalErrorWith_Unwrap(t *testing.T) {
	myErr := myError("myError")
	wrErr := fmt.Errorf("wrapped: %w", myErr) // double wrap
	twerr := twirp.InternalErrorWith(wrErr)

	if !errors.Is(twerr, myErr) {
		t.Errorf("expected errors.Is to match the error wrapped by twirp.InternalErrorWith")
	}

	var errTarget myError
	if !errors.As(twerr, &errTarget) {
		t.Errorf("expected errors.As to match the error wrapped by twirp.InternalErrorWith")
	}
	if errTarget.Error() != myErr.Error() {
		t.Errorf("invalid value for errTarget.Error(). have=%q, want=%q", errTarget.Error(), myErr.Error())
	}
}

type errorResponeWriter struct {
	*httptest.ResponseRecorder
}

func (errorResponeWriter) Write([]byte) (int, error) {
	return 0, errors.New("this is only a test")
}

type twerrJSON struct {
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
	Meta map[string]string `json:"meta,omitempty"`
}

func TestWriteError(t *testing.T) {
	resp := httptest.NewRecorder()
	twerr := twirp.NewError(twirp.Internal, "test middleware error")
	err := twirp.WriteError(resp, twerr)
	if err != nil {
		t.Errorf("got an error from WriteError when not expecting one: %s", err)
		return
	}

	twerrCode := twirp.ServerHTTPStatusFromErrorCode(twerr.Code())
	if resp.Code != twerrCode {
		t.Errorf("got wrong status. have=%d, want=%d", resp.Code, twerrCode)
		return
	}

	var gotTwerrJSON twerrJSON
	err = json.NewDecoder(resp.Body).Decode(&gotTwerrJSON)
	if err != nil {
		t.Errorf("got an error decoding response body: %s", err)
		return
	}

	if twirp.ErrorCode(gotTwerrJSON.Code) != twerr.Code() {
		t.Errorf("got wrong error code. have=%s, want=%s", gotTwerrJSON.Code, twerr.Code())
		return
	}

	if gotTwerrJSON.Msg != twerr.Msg() {
		t.Errorf("got wrong error message. have=%s, want=%s", gotTwerrJSON.Msg, twerr.Msg())
		return
	}

	errResp := &errorResponeWriter{ResponseRecorder: resp}

	// Writing again should error out as headers are being rewritten
	err = twirp.WriteError(errResp, twerr)
	if err == nil {
		t.Errorf("did not get error on write. have=nil, want=some error")
	}
}

func TestWriteError_WithNonTwirpError(t *testing.T) {
	resp := httptest.NewRecorder()
	nonTwerr := errors.New("not a twirp error")
	err := twirp.WriteError(resp, nonTwerr)
	if err != nil {
		t.Errorf("got an error from WriteError when not expecting one: %s", err)
		return
	}

	if resp.Code != 500 {
		t.Errorf("got wrong status. have=%d, want=%d", resp.Code, 500)
		return
	}

	var gotTwerrJSON twerrJSON
	err = json.NewDecoder(resp.Body).Decode(&gotTwerrJSON)
	if err != nil {
		t.Errorf("got an error decoding response body: %s", err)
		return
	}

	if twirp.ErrorCode(gotTwerrJSON.Code) != twirp.Internal {
		t.Errorf("got wrong error code. have=%s, want=%s", gotTwerrJSON.Code, twirp.Internal)
		return
	}

	if gotTwerrJSON.Msg != ""+nonTwerr.Error() {
		t.Errorf("got wrong error message. have=%s, want=%s", gotTwerrJSON.Msg, nonTwerr.Error())
		return
	}
}

// Test helpers

func assertTwirpError(t *testing.T, twerr twirp.Error, code twirp.ErrorCode, msg string) {
	t.Helper()
	if twerr.Code() != code {
		t.Errorf("wrong code. have=%q, want=%q", twerr.Code(), code)
	}
	if twerr.Msg() != msg {
		t.Errorf("wrong msg. have=%q, want=%q", twerr.Msg(), msg)
	}
}

func assertTwirpErrorMeta(t *testing.T, twerr twirp.Error, key string, value string) {
	t.Helper()
	if twerr.Meta(key) != value {
		t.Errorf("wrong meta. have=%q, want=%q", twerr.Meta(key), value)
	}
}
