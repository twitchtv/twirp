// Copyright 2019 Twitch Interactive, Inc.  All Rights Reserved.
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
	"log"
	"testing"
)

// testLogger creates a *log.Logger that writes log messages to the test's
// output. This makes log messages appear only if the test fails, and makes them
// align correctly for nested subtests.
func testLogger(t *testing.T) *log.Logger {
	return log.New(testWriter{t}, "", log.LstdFlags)
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(string(p))
	return len(p), nil
}
