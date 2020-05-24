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

// +build go1.13

package twirp_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/twitchtv/twirp"
)

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
