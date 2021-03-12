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
	"os"
	"os/signal"
	"syscall"

	server "github.com/twitchtv/twirp/example/internal"
	"github.com/twitchtv/twirp/example/rpc/example"
)

func main() {
	twirpServer, err := server.NewHaberdasherTwirpServer()
	if err != nil {
		log.Fatalf("Failed to create haberdasher server: %s", err.Error())
	}

	twirpHandler := example.NewHaberdasherServer(twirpServer)
	mux := http.NewServeMux()
	mux.Handle(example.HaberdasherPathPrefix, twirpHandler)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, int64(10<<20)) // 10  MiB max per request
		mux.ServeHTTP(w, r)
	})

	// This port can be in a config file at some point
	port := 8000
	appServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	shutdownCh := make(chan bool)
	go func() {
		shutdownSIGch := make(chan os.Signal, 1)
		signal.Notify(shutdownSIGch, syscall.SIGINT, syscall.SIGTERM)

		sig := <-shutdownSIGch

		log.Printf("Received shutdown signal request \n", "signal", sig)
		if err := appServer.Shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown server: %s", err.Error())
		}
		close(shutdownCh)
	}()

	log.Printf("Starting example twirp service on port %d \n", port)
	if err := appServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to listen and serve: %s", err.Error())
	}

	<-shutdownCh
	log.Println("Server successfully shutdown")
}
