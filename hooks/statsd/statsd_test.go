package statsd

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/internal/twirptest"
)

func TestSanitize(t *testing.T) {
	type test struct {
		in   string
		want string
	}
	cases := []test{
		{"", ""},
		{"SomeString", "SomeString"},
		{"with whitespace", "with_whitespace"},
		{"with.punctuation", "with_punctuation"},
		{"with世界unicode", "with__unicode"},
		{"with123numbers", "with123numbers"},
	}
	for _, c := range cases {
		have := sanitize(c.in)
		if have != c.want {
			t.Errorf("in=%q  have=%q  want=%q", c.in, have, c.want)
		}
	}
}

func TestTimingHooks(t *testing.T) {
	statter := &fakeStatter{}
	hooks := NewStatsdServerHooks(statter)

	server, client := serverAndClient(hooks)
	defer server.Close()

	_, err := client.MakeHat(context.Background(), &twirptest.Size{})
	if err != nil {
		t.Fatalf("twirptest Client err=%q", err)
	}

	expectedIncrements := []string{
		"twirp.total.requests",
		"twirp.MakeHat.requests",
		"twirp.total.responses",
		"twirp.MakeHat.responses",
		"twirp.status_codes.total.200",
		"twirp.status_codes.MakeHat.200",
	}
	for _, inc := range expectedIncrements {
		if !statter.receivedInc(inc) {
			t.Errorf("statter did not receive increment %q", inc)
		}
	}
	expectedTimers := []string{
		"twirp.all_methods.response",
		"twirp.MakeHat.response",
		"twirp.status_codes.all_methods.200",
		"twirp.status_codes.MakeHat.200",
	}
	for _, tim := range expectedTimers {
		if !statter.receivedTiming(tim) {
			t.Errorf("statter did not receive timing %q", tim)
		}
	}
}

func serverAndClient(hooks *twirp.ServerHooks) (*httptest.Server, twirptest.Haberdasher) {
	return twirptest.ServerAndClient(twirptest.NoopHatmaker(), hooks)
}

type increment struct {
	metric string
	val    int64
	rate   float32
}

type timing struct {
	metric string
	val    time.Duration
	rate   float32
}

type fakeStatter struct {
	incs  []increment
	times []timing
}

func (s *fakeStatter) Inc(metric string, val int64, rate float32) error {
	s.incs = append(s.incs, increment{metric, val, rate})
	return nil
}
func (s *fakeStatter) TimingDuration(metric string, val time.Duration, rate float32) error {
	s.times = append(s.times, timing{metric, val, rate})
	return nil
}

func (s *fakeStatter) receivedInc(metric string) bool {
	for _, i := range s.incs {
		if i.metric == metric {
			return true
		}
	}
	return false
}

func (s *fakeStatter) receivedTiming(metric string) bool {
	for _, t := range s.times {
		if t.metric == metric {
			return true
		}
	}
	return false
}
