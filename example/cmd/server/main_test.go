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
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/twitchtv/twirp/example"
)

func TestMain(m *testing.M) {
	go runServer()
	os.Exit(m.Run())
}

func runServer() {
	server := example.NewHaberdasherServer(&randomHaberdasher{}, nil)
	log.Fatal(http.ListenAndServe(":8080", server))
}

func newProtoClient() example.Haberdasher {
	return example.NewHaberdasherProtobufClient("http://localhost:8080", &http.Client{})
}

func newJSONClient() example.Haberdasher {
	return example.NewHaberdasherJSONClient("http://localhost:8080", &http.Client{})
}

type client struct {
	name string
	svc  example.Haberdasher
}

func clients() []client {
	return []client{
		{`Proto`, newProtoClient()},
		// {`JSON`, newJSONClient()},
	}
}

func compareErrors(got, expected error) error {
	if got == nil && expected == nil {
		return nil
	}
	if got == nil || expected == nil {
		return fmt.Errorf(`Expected err to be %#v, got %#v`, expected, got)
	}
	if got.Error() == expected.Error() {
		return nil
	}
	return fmt.Errorf(`Expected err to be %#v, got %#v`, expected, got)
}

func TestInvalidMakeHatsRequests(t *testing.T) {
	type testReq struct {
		name     string
		req      *example.MakeHatsReq
		expected error
	}
	testReqs := []testReq{
		{
			name:     `NegativeQuantity`,
			req:      &example.MakeHatsReq{Inches: 8, Quantity: -5},
			expected: errNegativeQuantity,
		},
		// // TooSmall is currently not being checked before the stream is returned so this would fail
		// {
		// 	name:     `TooSmall`,
		// 	req:      &example.MakeHatsReq{Inches: -5},
		// 	expected: errTooSmall,
		// },
	}

	for _, cc := range clients() {
		for _, re := range testReqs {
			t.Run(re.name+`/`+cc.name, func(t *testing.T) {
				hatStream, err := cc.svc.MakeHats(context.Background(), re.req)
				err = compareErrors(err, re.expected)
				if err != nil {
					t.Fatal(err)
				}
				if hatStream != nil {
					t.Fatalf(`expected hatStream to be nil, got %+v`, hatStream)
				}
			})
		}
	}
}

func TestMakeHatsTooSmall(t *testing.T) {
	for _, cc := range clients() {
		t.Run(cc.name, func(t *testing.T) {
			hatStream, err := cc.svc.MakeHats(
				context.Background(),
				&example.MakeHatsReq{Inches: -5, Quantity: 10},
			)
			err = compareErrors(err, nil)
			if err != nil {
				t.Fatal(err)
			}
			hatOrErr := <-hatStream
			err = compareErrors(hatOrErr.Err, errTooSmall)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMakeHatsLargeQuantities(t *testing.T) {
	type testReq struct {
		name string
		req  *example.MakeHatsReq
	}
	testReqs := []testReq{
		{name: `OneHundred`, req: &example.MakeHatsReq{Inches: 5, Quantity: 100}},
		{name: `OneThousand`, req: &example.MakeHatsReq{Inches: 5, Quantity: 1000}},
		{name: `TenThousand`, req: &example.MakeHatsReq{Inches: 5, Quantity: 10000}},
		{name: `OneHundredThousand`, req: &example.MakeHatsReq{Inches: 5, Quantity: 100000}},
		// {name: `OneMillion`, req:  &example.MakeHatsReq{Inches: 5, Quantity: 1000000}},
		// // OneMillion takes 6+ seconds if server is flushing after every message, <1sec if no flushing
	}

	for _, cc := range clients() {
		for _, re := range testReqs {
			t.Run(re.name+`/`+cc.name, func(t *testing.T) {
				reqSentAt := time.Now()
				hatStream, err := cc.svc.MakeHats(context.Background(), re.req)
				if err != nil {
					t.Fatalf(`MakeHats request failed: %#v (hatStream=%#v)`, err, hatStream)
				}
				ii := int32(0)
				for hatOrErr := range hatStream {
					if hatOrErr.Err != nil {
						t.Fatal(hatOrErr.Err)
					}
					ii++
				}
				if ii != re.req.Quantity {
					t.Fatalf(`Expected to receive %d hats, got %d`, re.req.Quantity, ii)
				}
				took := time.Now().Sub(reqSentAt)
				t.Logf(
					"Received %.1f kHats per second (%d hats in %f seconds)\n",
					float64(ii)/took.Seconds()/1000,
					ii, took.Seconds(),
				)
			})
		}
	}
}

func BenchmarkMakeHatsProto(b *testing.B) {
	benchmarkMakeHats(b, newProtoClient())
}

// func BenchmarkMakeHatsJSON(b *testing.B) {
// 	benchmarkMakeHats(b, newJSONClient())
// }

func benchmarkMakeHats(b *testing.B, cc example.Haberdasher) {
	reqSentAt := time.Now()
	hatStream, err := cc.MakeHats(
		context.Background(),
		&example.MakeHatsReq{Inches: 8, Quantity: int32(b.N)},
	)
	if err != nil {
		b.Fatal(err)
	}
	ii := 0
	for hatOrErr := range hatStream {
		if hatOrErr.Err != nil {
			b.Fatal(hatOrErr.Err)
		}
		ii++
	}
	if ii != b.N {
		b.Fatalf(`Expected to receive %d hats, got %d`, b.N, ii)
	}
	took := time.Now().Sub(reqSentAt)
	b.Logf(
		"Received %.1f kHats per second (%d hats in %f seconds)\n",
		float64(ii)/took.Seconds()/1000,
		ii, took.Seconds(),
	)
}
