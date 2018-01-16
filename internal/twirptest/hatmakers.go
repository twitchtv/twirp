package twirptest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/twitchtv/twirp"
)

type hatmaker func(ctx context.Context, s *Size) (*Hat, error)

func (h hatmaker) MakeHat(ctx context.Context, s *Size) (*Hat, error) { return h(ctx, s) }

// HaberdasherFunc is a convenience to convert a function into a Haberdasher service.
func HaberdasherFunc(f func(ctx context.Context, s *Size) (*Hat, error)) Haberdasher {
	return hatmaker(f)
}

// Always makes a blank hat.
func NoopHatmaker() Haberdasher {
	return hatmaker(func(context.Context, *Size) (*Hat, error) {
		return &Hat{}, nil
	})
}

// Makes a hat, as long as its the size they like
func PickyHatmaker(want int32) Haberdasher {
	return hatmaker(func(ctx context.Context, s *Size) (*Hat, error) {
		if s.Inches != want {
			return nil, twirp.InvalidArgumentError("Inches", "I can't make a hat that size!")
		}
		return &Hat{s.Inches, "blue", "top hat"}, nil
	})
}

// Makes a hat, but sure takes their time
func SlowHatmaker(dur time.Duration) Haberdasher {
	return hatmaker(func(ctx context.Context, s *Size) (*Hat, error) {
		time.Sleep(dur)
		return &Hat{s.Inches, "blue", "top hat"}, nil
	})
}

// Always errors.
func ErroringHatmaker(err error) Haberdasher {
	return hatmaker(func(ctx context.Context, s *Size) (*Hat, error) {
		return nil, err
	})
}

// Panics.
func PanickyHatmaker(msg string) Haberdasher {
	return hatmaker(func(ctx context.Context, s *Size) (*Hat, error) {
		panic(msg)
	})
}

// Returns nil, nil
func NilHatmaker() Haberdasher {
	return hatmaker(func(context.Context, *Size) (*Hat, error) {
		return nil, nil
	})
}

func ServerAndClient(h Haberdasher, hooks *twirp.ServerHooks) (*httptest.Server, Haberdasher) {
	s := httptest.NewServer(NewHaberdasherServer(h, hooks))
	c := NewHaberdasherProtobufClient(s.URL, http.DefaultClient)
	return s, c
}

func TwirpServerAndClient(hooks *twirp.ServerHooks) (*httptest.Server, Haberdasher) {
	return ServerAndClient(NoopHatmaker(), hooks)
}
