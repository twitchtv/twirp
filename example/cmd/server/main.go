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

type randomHaberdasher struct{}

func (h *randomHaberdasher) MakeHat(ctx context.Context, size *example.Size) (*example.Hat, error) {
	if size.Inches <= 0 {
		return nil, twirp.InvalidArgumentError("Inches", "I can't make a hat that small!")
	}
	return &example.Hat{
		Size:  size.Inches,
		Color: []string{"white", "black", "brown", "red", "blue"}[rand.Intn(4)],
		Name:  []string{"bowler", "baseball cap", "top hat", "derby"}[rand.Intn(3)],
	}, nil
}

func main() {
	hook := statsd.NewStatsdServerHooks(LoggingStatter{os.Stderr})
	server := example.NewHaberdasherServer(&randomHaberdasher{}, hook)
	log.Fatal(http.ListenAndServe(":8080", server))
}
