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

package json_serialization

import (
	bytes "bytes"
	"context"
	json "encoding/json"
	io "io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twitchtv/twirp"
)

type JSONSerializationService struct{}

func (h *JSONSerializationService) EchoJSON(ctx context.Context, req *Msg) (*Msg, error) {
	if req.AllEmpty {
		return &Msg{}, nil
	}

	return &Msg{
		Query:      req.Query,
		PageNumber: req.PageNumber,
		Hell:       req.Hell,
		Foobar:     req.Foobar,
		Snippets:   req.Snippets,
	}, nil
}

func TestJSONSerializationServiceWithDefaults(t *testing.T) {
	s := httptest.NewServer(NewJSONSerializationServer(&JSONSerializationService{}))
	defer s.Close()

	// Manual JSON request to get empty response
	// Response should include empty fields by default
	reqBody := bytes.NewBuffer([]byte(
		`{"allEmpty": true}`,
	))
	req, _ := http.NewRequest("POST", s.URL+"/twirp/JSONSerialization/EchoJSON", reqBody)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("manual EchoJSON err=%q", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("manual EchoJSON invalid status, have=%d, want=200", resp.StatusCode)
	}

	objmap := readJSONAsMap(t, resp.Body)
	for _, field := range []string{"query", "page_number", "hell", "foobar", "snippets", "all_empty"} {
		if _, ok := objmap[field]; !ok {
			t.Fatalf("expected JSON response to include field %q", field)
		}
	}

	// JSON Client
	client := NewJSONSerializationJSONClient(s.URL, http.DefaultClient)

	// Check empty fields
	msg, err := client.EchoJSON(context.Background(), &Msg{
		AllEmpty: true,
	})
	if err != nil {
		t.Fatalf("client.EchoJSON err=%q", err)
	}
	if have, want := msg.Query, ""; have != want {
		t.Fatalf("invalid msg.Query, have=%v, want=%v", have, want)
	}
	if have, want := msg.PageNumber, int32(0); have != want {
		t.Fatalf("invalid msg.PageNumber, have=%v, want=%v", have, want)
	}
	if have, want := msg.Hell, float64(0); have != want {
		t.Fatalf("invalid msg.Hell, have=%v, want=%v", have, want)
	}
	if have, want := msg.Foobar, Msg_FOO; have != want {
		t.Fatalf("invalid msg.Foobar, have=%v, want=%v", have, want)
	}
	if have, want := len(msg.Snippets), 0; have != want {
		t.Fatalf("invalid len(msg.Snippets), have=%v, want=%v", have, want)
	}

	// Check sending some values and reading the echo
	msg2, err := client.EchoJSON(context.Background(), &Msg{
		Query:      "my query",
		PageNumber: 33,
		Hell:       666.666,
		Foobar:     Msg_BAR,
		Snippets:   []string{"s1", "s2"},
	})
	if err != nil {
		t.Fatalf("client.DoJSON err=%q", err)
	}
	if have, want := msg2.Query, "my query"; have != want {
		t.Fatalf("invalid msg.Query, have=%v, want=%v", have, want)
	}
	if have, want := msg2.PageNumber, int32(33); have != want {
		t.Fatalf("invalid msg.PageNumber, have=%v, want=%v", have, want)
	}
	if have, want := msg2.Hell, 666.666; have != want {
		t.Fatalf("invalid msg.Hell, have=%v, want=%v", have, want)
	}
	if have, want := msg2.Foobar, Msg_BAR; have != want {
		t.Fatalf("invalid msg.Foobar, have=%v, want=%v", have, want)
	}
	if have, want := len(msg2.Snippets), 2; have != want {
		t.Fatalf("invalid len(msg.Snippets), have=%v, want=%v", have, want)
	}
	if have, want := msg2.Snippets[0], "s1"; have != want {
		t.Fatalf("invalid msg2.Snippets[0], have=%v, want=%v", have, want)
	}
}

func TestJSONSerializationServiceSkipDefaults(t *testing.T) {
	s := httptest.NewServer(
		NewJSONSerializationServer(
			&JSONSerializationService{},
			twirp.WithServerJSONSkipDefaults(true),
		),
	)
	defer s.Close()

	// Manual JSON request to get empty response.
	// Response should skip empty fields, in this case be completely empty
	reqBody := bytes.NewBuffer([]byte(
		`{"allEmpty": true}`,
	))
	req, _ := http.NewRequest("POST", s.URL+"/twirp/JSONSerialization/EchoJSON", reqBody)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("manual EchoJSON err=%q", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("manual EchoJSON invalid status, have=%d, want=200", resp.StatusCode)
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	if buf.String() != "{}" {
		t.Fatalf("manual EchoJSON response with JSONSkipDefaults expected to be '{}'. But it is=%q", buf.String())
	}

	// JSON Client
	client := NewJSONSerializationJSONClient(s.URL, http.DefaultClient)

	// Check empty fields
	msg, err := client.EchoJSON(context.Background(), &Msg{
		AllEmpty: true,
	})
	if err != nil {
		t.Fatalf("client.DoJSON err=%q", err)
	}
	if have, want := msg.Query, ""; have != want {
		t.Fatalf("invalid msg.Query, have=%v, want=%v", have, want)
	}
	if have, want := msg.PageNumber, int32(0); have != want {
		t.Fatalf("invalid msg.PageNumber, have=%v, want=%v", have, want)
	}
	if have, want := msg.Hell, float64(0); have != want {
		t.Fatalf("invalid msg.Hell, have=%v, want=%v", have, want)
	}
	if have, want := msg.Foobar, Msg_FOO; have != want {
		t.Fatalf("invalid msg.Foobar, have=%v, want=%v", have, want)
	}
	if have, want := len(msg.Snippets), 0; have != want {
		t.Fatalf("invalid len(msg.Snippets), have=%v, want=%v", have, want)
	}

	// Check sending some values and reading the echo
	msg2, err := client.EchoJSON(context.Background(), &Msg{
		Query:      "my query",
		PageNumber: 33,
		Hell:       666.666,
		Foobar:     Msg_BAR,
		Snippets:   []string{"s1", "s2"},
	})
	if err != nil {
		t.Fatalf("client.DoJSON err=%q", err)
	}
	if have, want := msg2.Query, "my query"; have != want {
		t.Fatalf("invalid msg.Query, have=%v, want=%v", have, want)
	}
	if have, want := msg2.PageNumber, int32(33); have != want {
		t.Fatalf("invalid msg.PageNumber, have=%v, want=%v", have, want)
	}
	if have, want := msg2.Hell, 666.666; have != want {
		t.Fatalf("invalid msg.Hell, have=%v, want=%v", have, want)
	}
	if have, want := msg2.Foobar, Msg_BAR; have != want {
		t.Fatalf("invalid msg.Foobar, have=%v, want=%v", have, want)
	}
	if have, want := len(msg2.Snippets), 2; have != want {
		t.Fatalf("invalid len(msg.Snippets), have=%v, want=%v", have, want)
	}
	if have, want := msg2.Snippets[0], "s1"; have != want {
		t.Fatalf("invalid msg2.Snippets[0], have=%v, want=%v", have, want)
	}
}

func TestJSONSerializationCamelCase(t *testing.T) {
	s := httptest.NewServer(
		NewJSONSerializationServer(
			&JSONSerializationService{},
			twirp.WithServerJSONCamelCaseNames(true),
		),
	)
	defer s.Close()

	reqBody := bytes.NewBuffer([]byte(`{"pageNumber": 123}`))
	req, _ := http.NewRequest("POST", s.URL+"/twirp/JSONSerialization/EchoJSON", reqBody)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("manual EchoJSON err=%q", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("manual EchoJSON invalid status, have=%d, want=200", resp.StatusCode)
	}

	objmap := readJSONAsMap(t, resp.Body)

	// response includes camelCase names
	for _, field := range []string{"query", "pageNumber", "hell", "foobar", "snippets", "allEmpty"} {
		if _, ok := objmap[field]; !ok {
			t.Fatalf("expected JSON response to include camelCase field %q", field)
		}
	}

	// response does not include original snake_case names
	for _, field := range []string{"page_number", "all_empty"} {
		if _, ok := objmap[field]; ok {
			t.Fatalf("expected JSON response to NOT include snake_case field %q", field)
		}
	}
}

//
// Test helpers
//

func readJSONAsMap(t *testing.T, body io.Reader) map[string]json.RawMessage {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(body)
	var objmap map[string]json.RawMessage
	err := json.Unmarshal(buf.Bytes(), &objmap)
	if err != nil {
		t.Fatalf("json.Unmarshal err=%q", err)
	}
	return objmap
}
