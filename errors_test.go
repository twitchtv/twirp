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
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

func TestWithMetaRaces(t *testing.T) {
	err := NewError(Internal, "msg")
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
	twerr := InternalErrorWith(rootCause)
	cause := pkgerrors.Cause(twerr)
	if cause != rootCause {
		t.Errorf("got wrong cause for err. have=%q, want=%q", cause, rootCause)
	}
}

func TestWrapError(t *testing.T) {
	rootCause := errors.New("cause")
	twerr := NewError(NotFound, "it ain't there")
	err := WrapError(twerr, rootCause)
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
	twerr := InternalErrorWith(wrErr)

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

func TestWriteError(t *testing.T) {
	resp := httptest.NewRecorder()
	twerr := NewError(Internal, "test middleware error")
	err := WriteError(resp, twerr)
	if err != nil {
		t.Errorf("got an error from WriteError when not expecting one: %s", err)
		return
	}

	twerrCode := ServerHTTPStatusFromErrorCode(twerr.Code())
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

	if ErrorCode(gotTwerrJSON.Code) != twerr.Code() {
		t.Errorf("got wrong error code. have=%s, want=%s", gotTwerrJSON.Code, twerr.Code())
		return
	}

	if gotTwerrJSON.Msg != twerr.Msg() {
		t.Errorf("got wrong error message. have=%s, want=%s", gotTwerrJSON.Msg, twerr.Msg())
		return
	}

	errResp := &errorResponeWriter{ResponseRecorder: resp}

	// Writing again should error out as headers are being rewritten
	err = WriteError(errResp, twerr)
	if err == nil {
		t.Errorf("did not get error on write. have=nil, want=some error")
	}
}

func TestWriteError_WithNonTwirpError(t *testing.T) {
	resp := httptest.NewRecorder()
	nonTwerr := errors.New("not a twirp error")
	err := WriteError(resp, nonTwerr)
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

	if ErrorCode(gotTwerrJSON.Code) != Internal {
		t.Errorf("got wrong error code. have=%s, want=%s", gotTwerrJSON.Code, Internal)
		return
	}

	if gotTwerrJSON.Msg != ""+nonTwerr.Error() {
		t.Errorf("got wrong error message. have=%s, want=%s", gotTwerrJSON.Msg, nonTwerr.Error())
		return
	}
}
