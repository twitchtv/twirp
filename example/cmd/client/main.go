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
