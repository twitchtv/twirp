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
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/example"
	"github.com/twitchtv/twirp/hooks/statsd"
)

type randomHaberdasher struct{}

var (
	errTooSmall         = twirp.InvalidArgumentError("Inches", "I can't make hats that small!")
	errNegativeQuantity = twirp.InvalidArgumentError("Quantity", "I can't make a negative quantity of hats!")
)

func newRandomHat(inches int32) (*example.Hat, error) {
	if inches <= 0 {
		return nil, errTooSmall
	}
	return &example.Hat{
		Size:  inches,
		Color: []string{"white", "black", "brown", "red", "blue"}[rand.Intn(5)],
		Name:  []string{"bowler", "baseball cap", "top hat", "derby"}[rand.Intn(4)],
	}, nil
}

func (h *randomHaberdasher) MakeHat(ctx context.Context, size *example.Size) (*example.Hat, error) {
	return newRandomHat(size.Inches)
}

func (h *randomHaberdasher) MakeHats(ctx context.Context, req *example.MakeHatsReq) (example.HatStream, error) {
	if req.Quantity < 0 {
		return nil, errNegativeQuantity
	}
	// Normally we'd validate Inches here as well, but we let it fall through to error on newRandomHat to demonstrate mid-stream errors
	// if req.Inches <= 0 {
	// 	return nil, errTooSmall
	// }

	ch := make(chan example.HatOrError, 100) // NB: the size of this buffer can make a big difference!
	go func() {
		for ii := int32(0); ii < req.Quantity; ii++ {
			hat, err := newRandomHat(req.Inches)
			select {
			case <-ctx.Done():
				return
			case ch <- example.HatOrError{Hat: hat, Err: err}:
			}
		}
		close(ch)
	}()
	return example.NewHatStream(ch), nil
}

func main() {
	hook := statsd.NewStatsdServerHooks(LoggingStatter{os.Stderr})
	server := example.NewHaberdasherServer(&randomHaberdasher{}, hook)
	log.Fatal(http.ListenAndServe(":8080", server))
}
