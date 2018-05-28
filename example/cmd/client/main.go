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
	"io"
	"log"
	"net/http"
	"time"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/example"
)

func main() {
	client := example.NewHaberdasherProtobufClient("http://localhost:8080", &http.Client{})

	var (
		hat *example.Hat
		err error
	)

	//
	// Call the MakeHat rpc
	//
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
		break
	}
	log.Println(`Response from MakeHat:`)
	log.Printf("\t%+v\n", hat)

	//
	// Call the MakeHats streaming rpc
	//
	const (
		printEvery = 50000
		quantity   = int32(300000)
	)
	reqSentAt := time.Now()
	hatStream, err := client.MakeHats(
		context.Background(),
		&example.MakeHatsReq{Inches: 12, Quantity: quantity},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Response from MakeHats:\n")
	ii := 1
	printResults := func() {
		took := time.Now().Sub(reqSentAt)
		khps := float64(ii-1) / took.Seconds() / 1000
		log.Printf("Received %.1f kHats per second (%d hats in %f seconds)\n", khps, ii-1, took.Seconds())
	}
	for ; true; ii++ { // Receive all the hats
		hat, err = hatStream.Next(context.Background())
		if err != nil {
			if err == io.EOF {
				break
			}
			printResults()
			log.Fatal(err)
		}
		if ii%printEvery == 0 {
			khps := float64(ii) / time.Now().Sub(reqSentAt).Seconds() / 1000
			log.Printf("\t[%4.1f khps] %6d %+v\n", khps, ii, hat)
		}
	}
	printResults()
}
