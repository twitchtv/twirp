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

package twirptest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/twitchtv/twirp"
)

func TestClientContextCanceled(t *testing.T) {
	// Context that is already canceled.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Make request, should immediately fail because the context is canceled.
	protoCli := NewHaberdasherProtobufClient("", &http.Client{})
	_, err := protoCli.MakeHat(ctx, &Size{})

	// The failure is a twirp internal error
	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error, have=%T", err)
	}
	if twerr.Code() != twirp.Internal {
		t.Fatalf("expected twirp.Error Code to be internal, have=%q", twerr.Code())
	}

	// But the context.Canceled error can be identified easily with errors.Is
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected errors.Is(err, context.Canceled) to match, but it didn't")
	}
}

func TestClientErrorsCanBeUnwrapped(t *testing.T) {
	rootErr := fmt.Errorf("some root cause")
	httpClient := &http.Client{
		Transport: &failingTransport{rootErr},
	}

	client := NewHaberdasherJSONClient("", httpClient)
	_, err := client.MakeHat(context.Background(), &Size{Inches: 1})
	if err == nil {
		t.Errorf("JSON MakeHat err is unexpectedly nil")
	}
	if !errors.Is(err, rootErr) {
		t.Fatalf("expected errors.Is(err, rootErr) to match, but it didn't")
	}
}
