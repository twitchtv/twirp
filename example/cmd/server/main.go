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
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/example"
	"github.com/twitchtv/twirp/hooks/statsd"
)

func newRandomHat(inches int32) *example.Hat {
	return &example.Hat{
		Size:  inches,
		Color: []string{"white", "black", "brown", "red", "blue"}[rand.Intn(4)],
		Name:  []string{"bowler", "baseball cap", "top hat", "derby"}[rand.Intn(3)],
	}
}

type randomHatStream struct{ i, q int32 }

func (hs *randomHatStream) Next(ctx context.Context) (*example.Hat, error) {
	if hs.q == 0 {
		return nil, io.EOF
	}
	hs.q--
	time.Sleep(300 * time.Millisecond)
	return newRandomHat(hs.i), nil
}

func (hs *randomHatStream) End(err error) {
	// TODO: something?
}

type randomHaberdasher struct{}

func (h *randomHaberdasher) MakeHat(ctx context.Context, size *example.Size) (*example.Hat, error) {
	if size.Inches <= 0 {
		return nil, twirp.InvalidArgumentError("Inches", "I can't make a hat that small!")
	}
	return newRandomHat(size.Inches), nil
}

func (h *randomHaberdasher) MakeHats(ctx context.Context, req *example.MakeHatsReq) (example.HatStream, error) {
	if req.Inches <= 0 {
		return nil, twirp.InvalidArgumentError("Inches", "I can't make hats that small!")
	}
	if req.Quantity < 0 {
		return nil, twirp.InvalidArgumentError("Quantity", "I can't make a negative quantity of hats!")
	}
	return &randomHatStream{i: req.Inches, q: req.Quantity}, nil
}

func main() {
	hook := statsd.NewStatsdServerHooks(LoggingStatter{os.Stderr})
	server := example.NewHaberdasherServer(&randomHaberdasher{}, hook)
	log.Fatal(http.ListenAndServe(":8080", server))
}
