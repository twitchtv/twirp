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

package statsd

import (
	"context"
	"strings"
	"time"

	"github.com/twitchtv/twirp"
)

var reqStartTimestampKey = new(int)

func markReqStart(ctx context.Context) context.Context {
	return context.WithValue(ctx, reqStartTimestampKey, time.Now())
}

func getReqStart(ctx context.Context) (time.Time, bool) {
	t, ok := ctx.Value(reqStartTimestampKey).(time.Time)
	return t, ok
}

type Statter interface {
	Inc(metric string, val int64, rate float32) error
	TimingDuration(metric string, val time.Duration, rate float32) error
}

// NewStatsdServerHooks provides a twirp.ServerHooks struct which
// sends data to statsd.
func NewStatsdServerHooks(stats Statter) *twirp.ServerHooks {
	hooks := &twirp.ServerHooks{}
	// RequestReceived: inc twirp.total.req_recv
	hooks.RequestReceived = func(ctx context.Context) (context.Context, error) {
		ctx = markReqStart(ctx)
		stats.Inc("twirp.total.requests", 1, 1.0)
		return ctx, nil
	}

	// RequestRouted: inc twirp.<method>.req_recv
	hooks.RequestRouted = func(ctx context.Context) (context.Context, error) {
		method, ok := twirp.MethodName(ctx)
		if !ok {
			return ctx, nil
		}
		stats.Inc("twirp."+sanitize(method)+".requests", 1, 1.0)
		return ctx, nil
	}

	// ResponseSent:
	// - inc twirp.total.response
	// - inc twirp.<method>.response
	// - inc twirp.by_code.total.<code>.response
	// - inc twirp.by_code.<method>.<code>.response
	// - time twirp.total.response
	// - time twirp.<method>.response
	// - time twirp.by_code.total.<code>.response
	// - time twirp.by_code.<method>.<code>.response
	hooks.ResponseSent = func(ctx context.Context) {
		// Three pieces of data to get, none are guaranteed to be present:
		// - time that the request started
		// - method that was called
		// - status code of response
		var (
			start  time.Time
			method string
			status string

			haveStart  bool
			haveMethod bool
			haveStatus bool
		)

		start, haveStart = getReqStart(ctx)
		method, haveMethod = twirp.MethodName(ctx)
		status, haveStatus = twirp.StatusCode(ctx)

		method = sanitize(method)
		status = sanitize(status)

		stats.Inc("twirp.total.responses", 1, 1.0)

		if haveMethod {
			stats.Inc("twirp."+method+".responses", 1, 1.0)
		}
		if haveStatus {
			stats.Inc("twirp.status_codes.total."+status, 1, 1.0)
		}
		if haveMethod && haveStatus {
			stats.Inc("twirp.status_codes."+method+"."+status, 1, 1.0)
		}

		if haveStart {
			dur := time.Now().Sub(start)
			stats.TimingDuration("twirp.all_methods.response", dur, 1.0)

			if haveMethod {
				stats.TimingDuration("twirp."+method+".response", dur, 1.0)
			}
			if haveStatus {
				stats.TimingDuration("twirp.status_codes.all_methods."+status, dur, 1.0)
			}
			if haveMethod && haveStatus {
				stats.TimingDuration("twirp.status_codes."+method+"."+status, dur, 1.0)
			}
		}
	}
	return hooks
}

func sanitize(s string) string {
	return strings.Map(sanitizeRune, s)
}

func sanitizeRune(r rune) rune {
	switch {
	case 'a' <= r && r <= 'z':
		return r
	case '0' <= r && r <= '9':
		return r
	case 'A' <= r && r <= 'Z':
		return r
	default:
		return '_'
	}
}
