package main

import (
	"context"
	"fmt"
	"io"
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
	server := example.NewHaberdasherServer(&randomHaberdasher{quiet: true}, nil)
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
	c    example.Haberdasher
}

func clients() []client {
	return []client{
		{`Proto`, newProtoClient()},
		{`JSON`, newJSONClient()},
	}
}

func compareErrors(got, expected error) error {
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
			name:     `TooSmall`,
			req:      &example.MakeHatsReq{Inches: -5},
			expected: errTooSmall,
		},
		{
			name:     `NegativeQuantity`,
			req:      &example.MakeHatsReq{Inches: 8, Quantity: -5},
			expected: errNegativeQuantity,
		},
	}

	for _, cc := range clients() {
		for _, re := range testReqs {
			t.Run(re.name+cc.name, func(t *testing.T) {
				hatStream, err := cc.c.MakeHats(context.Background(), re.req)
				if err != nil {
					t.Fatalf(`MakeHats request failed: %#v`, err)
				}
				_, err = hatStream.Next(context.Background())
				err = compareErrors(err, re.expected)
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestMakeHatsPerf(t *testing.T) {
	type testReq struct {
		name string
		req  *example.MakeHatsReq
	}
	testReqs := []testReq{
		{
			name: `OneHundred`,
			req:  &example.MakeHatsReq{Inches: 5, Quantity: 100},
		},
		{
			name: `OneThousand`,
			req:  &example.MakeHatsReq{Inches: 5, Quantity: 1000},
		},
		{
			name: `TenThousand`,
			req:  &example.MakeHatsReq{Inches: 5, Quantity: 10000},
		},
		{
			name: `OneHundredThousand`,
			req:  &example.MakeHatsReq{Inches: 5, Quantity: 100000},
		},
		// // OneMillion takes 6+ seconds if server is Flush()ing after every message, <1sec if no flushing
		// {
		// 	name: `OneMillion`,
		// 	req:  &example.MakeHatsReq{Inches: 5, Quantity: 1000000},
		// },
	}

	for _, cc := range clients() {
		for _, re := range testReqs {
			t.Run(re.name+cc.name, func(t *testing.T) {
				reqSentAt := time.Now()
				hatStream, err := cc.c.MakeHats(context.Background(), re.req)
				if err != nil {
					t.Fatalf(`MakeHats request failed: %#v (hatStream=%#v)`, err, hatStream)
				}
				ii := int32(0)
				for ; true; ii++ {
					_, err = hatStream.Next(context.Background())
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
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

func BenchmarkMakeHatsJSON(b *testing.B) {
	benchmarkMakeHats(b, newJSONClient())
}

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
	for ; true; ii++ {
		_, err = hatStream.Next(context.Background())
		if err != nil {
			if err == io.EOF {
				break
			}
			b.Fatal(err)
		}
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
