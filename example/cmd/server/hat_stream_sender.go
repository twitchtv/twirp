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

// This file is an alternate implementation of the twirp-provided stream
// constructor originally proposed in the "As the sender" section of
// https://github.com/twitchtv/twirp/issues/70#issuecomment-361365458

import (
	"context"
	"io"

	"github.com/twitchtv/twirp/example"
)

type HatOrError struct {
	hat *example.Hat
	err error
}

func NewHatStream(ch chan HatOrError) *hatStreamSender {
	return &hatStreamSender{ch: ch}
}

type hatStreamSender struct {
	ch <-chan HatOrError
}

func (hs *hatStreamSender) Next(ctx context.Context) (*example.Hat, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case v, open := <-hs.ch:
		if !open {
			return nil, io.EOF
		}
		if v.err != nil {
			return nil, v.err
		}
		return v.hat, nil
	}
}

func (hs *hatStreamSender) End(err error) {} // Should anything go here?
