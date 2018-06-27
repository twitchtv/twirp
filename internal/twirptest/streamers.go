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

package twirptest

import "context"

type streamer struct {
	err error

	download chan RespOrError
}

func (s *streamer) Transact(ctx context.Context, in *Req) (*Resp, error) {
	if s.err != nil {
		return nil, s.err
	}
	return nil, nil
}

func (s *streamer) Upload(ctx context.Context, in <-chan ReqOrError) (*Resp, error) {
	if s.err != nil {
		return nil, s.err
	}
	return nil, nil
}

func (s *streamer) Download(ctx context.Context, in *Req) (<-chan RespOrError, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.download, nil
}

func (s *streamer) Communicate(ctx context.Context, in <-chan ReqOrError) (<-chan RespOrError, error) {
	if s.err != nil {
		return nil, s.err
	}
	return nil, nil
}
