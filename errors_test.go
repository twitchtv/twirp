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
	"fmt"
	"sync"
	"testing"

	"github.com/pkg/errors"
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

func TestErrorCause(t *testing.T) {
	rootCause := fmt.Errorf("this is only a test")
	twerr := InternalErrorWith(rootCause)
	cause := errors.Cause(twerr)
	if cause != rootCause {
		t.Errorf("got wrong cause for err. have=%q, want=%q", cause, rootCause)
	}
}
