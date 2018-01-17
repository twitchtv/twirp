package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/example"
)

func main() {
	client := example.NewHaberdasherJSONClient("http://localhost:8080", &http.Client{})

	var (
		hat *example.Hat
		err error
	)
	for i := 0; i < 5; i++ {
		hat, err = client.MakeHat(context.Background(), &example.Size{Inches: 12})
		if err != nil {
			if twerr, ok := err.(twirp.Error); ok {
				if twerr.Meta("retryable") != "" {
					// Log the error and go again.
					log.Printf("got error %q, retrying", twerr)
					continue
				}
			}
			// This was some fatal error!
			log.Fatal(err)
		}
	}
	fmt.Printf("%+v", hat)
}
