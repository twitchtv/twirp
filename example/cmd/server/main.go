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

func newRandomHat(inches int32) *example.Hat {
	return &example.Hat{
		Size:  inches,
		Color: []string{"white", "black", "brown", "red", "blue"}[rand.Intn(5)],
		Name:  []string{"bowler", "baseball cap", "top hat", "derby"}[rand.Intn(4)],
	}
}

type randomHaberdasher struct{ quiet bool }

var (
	errTooSmall         = twirp.InvalidArgumentError("Inches", "I can't make hats that small!")
	errNegativeQuantity = twirp.InvalidArgumentError("Quantity", "I can't make a negative quantity of hats!")
)

func (h *randomHaberdasher) MakeHat(ctx context.Context, size *example.Size) (*example.Hat, error) {
	if size.Inches <= 0 {
		return nil, errTooSmall
	}
	return newRandomHat(size.Inches), nil
}

func (h *randomHaberdasher) MakeHats(ctx context.Context, req *example.MakeHatsReq) (example.HatStream, error) {
	if req.Inches <= 0 {
		return nil, errTooSmall
	}
	if req.Quantity < 0 {
		return nil, errNegativeQuantity
	}
	return newRandomHatStream(req.Inches, req.Quantity, h.quiet), nil
}

func main() {
	hook := statsd.NewStatsdServerHooks(LoggingStatter{os.Stderr})
	server := example.NewHaberdasherServer(&randomHaberdasher{}, hook)
	log.Fatal(http.ListenAndServe(":8080", server))
}
