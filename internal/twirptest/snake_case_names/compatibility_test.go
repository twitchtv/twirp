// Copyright 2019 Twitch Interactive, Inc.  All Rights Reserved.
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

package snake_case_names

import (
	context "context"
	http "net/http"
	"net/http/httptest"
	"testing"
)

type HaberdasherService struct{}

func (h *HaberdasherService) MakeHatV1(ctx context.Context, s *MakeHatArgsV1_SizeV1) (*MakeHatArgsV1_HatV1, error) {
	return &MakeHatArgsV1_HatV1{
		Size: s.Inches,
	}, nil
}

// When the proto definition contains service and/or method names with underscores (not following proto naming
// best practices), Go clients will mistakenly convert routes into it's CamelCased versions, but clients in other
// languages may keep the literal casing of the routes. This test makes a go client that would send CamelCased routes
// and checks that the generated Go server remains backwards compatible and is able to handle those routes.
func TestServiceMethodNamesCamelCase(t *testing.T) {
	s := httptest.NewServer(NewHaberdasherV1Server(&HaberdasherService{}, nil))
	defer s.Close()

	client := NewHaberdasherV1ProtobufClient(s.URL, http.DefaultClient)

	hat, err := client.MakeHatV1(context.Background(), &MakeHatArgsV1_SizeV1{Inches: 1})
	if err != nil {
		t.Fatalf("go protobuf client err=%q", err)
	}
	if hat.Size != 1 {
		t.Errorf("wrong hat size returned")
	}
}

type compatibilityTestClient struct {
	client *http.Client
}

func (t compatibilityTestClient) Do(req *http.Request) (*http.Response, error) {
	req.URL.Path = "/twirp/twirp.internal.twirptest.snake_case_names.Haberdasher_v1/MakeHat_v1"
	return t.client.Do(req)
}

// When the proto definition contains service and/or method names with underscores (not following proto naming
// best practices), Go clients will mistakenly convert routes into it's CamelCased versions, but clients in other
// languages may keep the literal casing of the routes. This test makes a fake client that would send literal routes
// and checks that the generated Go server is still able to handle those routes.
func TestServiceMethodNamesUnderscores(t *testing.T) {
	s := httptest.NewServer(NewHaberdasherV1Server(&HaberdasherService{}, nil))
	defer s.Close()

	client := NewHaberdasherV1ProtobufClient(s.URL, compatibilityTestClient{client: http.DefaultClient})
	hat, err := client.MakeHatV1(context.Background(), &MakeHatArgsV1_SizeV1{Inches: 1})
	if err != nil {
		t.Fatalf("compatible protobuf client err=%q", err)
	}
	if hat.Size != 1 {
		t.Errorf("wrong hat size returned")
	}

}
