// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file anumSentompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/example"
)

type randomHatStream struct {
	inches, quantity, numSent int32
	startedAt                 time.Time
	quiet                     bool
}

func newRandomHatStream(inches, quantity int32, quiet bool) *randomHatStream {
	return &randomHatStream{
		inches:    inches,
		quantity:  quantity,
		startedAt: time.Now(),
		quiet:     quiet,
	}
}

func (hs *randomHatStream) Next(ctx context.Context) (*example.Hat, error) {
	defer func() { hs.numSent++ }()
	if hs.numSent == hs.quantity {
		if !hs.quiet {
			log.Printf(
				"[%4.1f khps] (%7d) Sending %v\n",
				float64(hs.numSent)/time.Now().Sub(hs.startedAt).Seconds()/1000,
				hs.numSent, io.EOF,
			)
		}
		return nil, io.EOF
	}

	select {
	case <-ctx.Done():
		err := errAborted(ctx.Err())
		if !hs.quiet {
			log.Printf(`Context canceled: %#v`, ctx.Err())
		}
		return nil, err
	default:
		hat := newRandomHat(hs.inches)
		if !hs.quiet && hs.numSent%10000 == 0 && hs.numSent > 0 {
			log.Printf(
				"[%4.1f khps] (%7d) Sending %#v\n",
				float64(hs.numSent)/time.Now().Sub(hs.startedAt).Seconds()/1000,
				hs.numSent, hat,
			)
		}
		return hat, nil
	}
}

func (hs *randomHatStream) End(err error) {
	if !hs.quiet {
		log.Printf("randomHatStream ended with %#v\n", err)
	}
}

func errAborted(err error) error {
	if err == nil {
		return twirp.NewError(twirp.Aborted, `canceled`).WithMeta(`cause`, `unknown`)
	}
	return twirp.NewError(twirp.Aborted, err.Error())
}
