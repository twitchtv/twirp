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
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/twitchtv/twirp/internal/contextkeys"
)

func TestWithHTTPRequestHeaders(t *testing.T) {
	// 1. Test for error conditions.
	// Providing any one of these headers to "WithHTTPRequestHeaders" should return an error
	var ForbiddenHeaders = []string{"Accept", "Content-Type", "Twirp-Version"}
	for _, v := range ForbiddenHeaders {
		header := make(http.Header)
		header.Add(v, "")
		_, err := WithHTTPRequestHeaders(context.Background(), header)

		if err == nil {
			t.Fatalf("Providing the header name '%s' to WithHTTPRequestHeaders should return an error", v)
		}
	}

	// 2. Ensure that valid headers are added to the context
	type Test struct {
		Headers map[string][]string
		Context context.Context
		Output  context.Context
	}

	var background = context.Background()
	// These are test cases which should not return errors
	tests := []Test{
		{
			Headers: map[string][]string{
				"Authorization": []string{"Bearer qLtKDf9lS5r7QBgWgpQw"},
			},
			Output: context.WithValue(background, contextkeys.RequestHeaderKey, http.Header(map[string][]string{
				"Authorization": []string{"Bearer qLtKDf9lS5r7QBgWgpQw"},
			})),
		},
		{
			Headers: map[string][]string{
				"Authorization":  []string{"Bearer qLtKDf9lS5r7QBgWgpQw"},
				"Example-Header": []string{"e8L4ofltdHT6aTFcuKjm"},
			},
			Output: context.WithValue(background, contextkeys.RequestHeaderKey, http.Header(map[string][]string{
				"Authorization":  []string{"Bearer qLtKDf9lS5r7QBgWgpQw"},
				"Example-Header": []string{"e8L4ofltdHT6aTFcuKjm"},
			})),
		},
		{
			Headers: map[string][]string{
				"Authorization":   []string{"Bearer qLtKDf9lS5r7QBgWgpQw"},
				"X-Custom-Header": []string{"TiySVVYq57xDI7NoTFpP  C8oB9xYl2GBHUmkNulsp"},
			},
			Output: context.WithValue(background, contextkeys.RequestHeaderKey, http.Header(map[string][]string{
				"Authorization":   []string{"Bearer qLtKDf9lS5r7QBgWgpQw"},
				"X-Custom-Header": []string{"TiySVVYq57xDI7NoTFpP  C8oB9xYl2GBHUmkNulsp"},
			})),
		},
		{
			Headers: map[string][]string{
				"Authorization":   []string{"Bearer qLtKDf9lS5r7QBgWgpQw"},
				"X-Custom-Header": []string{"{\"e8L4ofltdHT6aTFcuKjm\":\"xQ2QPwziDKXYU93egsW4\",\"C8oB9xYl\":{\"e7i\":\"FgvPWtHPv55fAEVdQ\"}}"},
			},
			Output: context.WithValue(background, contextkeys.RequestHeaderKey, http.Header(map[string][]string{
				"Authorization":   []string{"Bearer qLtKDf9lS5r7QBgWgpQw"},
				"X-Custom-Header": []string{"{\"e8L4ofltdHT6aTFcuKjm\":\"xQ2QPwziDKXYU93egsW4\",\"C8oB9xYl\":{\"e7i\":\"FgvPWtHPv55fAEVdQ\"}}"},
			})),
		},
	}
	for i, v := range tests {
		ctx, err := WithHTTPRequestHeaders(v.Context, v.Headers)
		if err != nil {
			t.Errorf("Test case %d failed. Error: %s", i, err.Error())
		}
		received := ctx.Value(contextkeys.RequestHeaderKey)
		expected := v.Output.Value(contextkeys.RequestHeaderKey)
		if cmp.Equal(received, expected) != true {
			t.Errorf("Test case %d failed. Input does not match expected output.\n%s", i, cmp.Diff(received, expected))
		}
	}
}

func TestHTTPRequestHeaders(t *testing.T) {
	headers := []map[string]string{
		map[string]string{
			"Authorization": "Bearer qLtKDf9lS5r7QBgWgpQw",
		},
		map[string]string{
			"Authorization":  "Bearer qLtKDf9lS5r7QBgWgpQw",
			"Example-Header": "e8L4ofltdHT6aTFcuKjm",
		},
		map[string]string{
			"Authorization":   "Bearer qLtKDf9lS5r7QBgWgpQw",
			"X-Custom-Header": "TiySVVYq57xDI7NoTFpP  C8oB9xYl2GBHUmkNulsp",
		},
		map[string]string{
			"Authorization":   "Bearer qLtKDf9lS5r7QBgWgpQw",
			"X-Custom-Header": "{\"e8L4ofltdHT6aTFcuKjm\":\"xQ2QPwziDKXYU93egsW4\",\"C8oB9xYl\":{\"e7i\":\"FgvPWtHPv55fAEVdQ\"}}",
		},
	}

	tests := make([]http.Header, len(headers))

	for i, v := range headers {
		for key, value := range v {
			tests[i] = make(http.Header)
			tests[i].Set(key, value)
		}
	}

	for i, v := range tests {
		ctx, err := WithHTTPRequestHeaders(context.Background(), v)
		if err != nil {
			t.Errorf("Failed to create context for test case %d failed. Error: %s", i, err.Error())
		}
		received, ok := HTTPRequestHeaders(ctx)
		if !ok {
			t.Errorf("Test case %d failed; reading request headers failed. (returned 'false' for 'ok')", i)
		}
		expected := v
		if cmp.Equal(received, expected) != true {
			t.Errorf("Test case %d failed. Input does not match expected output.\n%s", i, cmp.Diff(received, expected))
		}
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.RequestHeaderKey, http.Header{"Example-Authorization": []string{"testing"}})
	ctx = context.WithValue(ctx, contextkeys.ServiceNameKey, "example.service")
	ctx = context.WithValue(ctx, contextkeys.PackageNameKey, "example")
	ctx = context.WithValue(ctx, contextkeys.MethodNameKey, "ex")

	header, ok := HTTPRequestHeaders(ctx)
	if !ok {
		t.Error("HTTPRequestHeaders returned false with context with multiple values")
	}
	if header == nil {
		t.Error("When providing a context with multiple values, HTTPRequestHeaders returned nil")
	}
	if cmp.Equal(header, map[string][]string{}) {
		t.Error("When providing a context with multiple values, HTTPRequestHeaders returned an empty map")
	}
}
