---
id: intro
title: Meet Twirp!
sidebar_label: Overview
---

Twirp is a simple RPC framework built on
[protobuf](https://developers.google.com/protocol-buffers/). You define a
service in a `.proto` specification file, then Twirp will _generate_ servers and
clients for that service. It's your job to fill in the "business logic" that
powers the server, and then generated clients can consume your service straight
away.

This doc is an overview of how you use Twirp - how you interact with it, what
you write, and what it generates.

## Making a Twirp Service

To make a Twirp service:

  1. Define your service in a **Proto** file.
  2. Use the `protoc` command to generate go code from the **Proto** file, it
     will generate an **interface**, a **client** and some **server utils** (to
     easily start an http listener).
  3. Implement the generated **interface** to implement the service.

For example, a HelloWorld **Proto** file:

```protobuf
syntax = "proto3";
package twitch.twirp.example.helloworld;
option go_package = "helloworld";

service HelloWorld {
  rpc Hello(HelloReq) returns (HelloResp);
}

message HelloReq {
  string subject = 1;
}

message HelloResp {
  string text = 1;
}
```

From which Twirp can auto-generate this **interface** (running the `protoc` command):

```go
type HelloWorld interface {
	Hello(context.Context, *HelloReq) (*HelloResp, error)
}
```

You provide the **implementation**:

```go
package main

import (
	"context"
	"net/http"

	pb "github.com/twitchtv/twirp-example/rpc/helloworld"
)

type HelloWorldServer struct{}

func (s *HelloWorldServer) Hello(ctx context.Context, req *pb.HelloReq) (*pb.HelloResp, error) {
	return &pb.HelloResp{Text: "Hello " + req.Subject}, nil
}

// Run the implementation in a local server
func main() {
	twirpHandler := pb.NewHelloWorldServer(&HelloWorldServer{}, nil)
	// You can use any mux you like - NewHelloWorldServer gives you an http.Handler.
	mux := http.NewServeMux()
	// The generated code includes a const, <ServiceName>PathPrefix, which
	// can be used to mount your service on a mux.
	mux.Handle(pb.HelloWorldPathPrefix, twirpHandler)
	http.ListenAndServe(":8080", mux)
}
```

And voila! Now you can just use the auto-generated **Client** to make remote calls to your new service:

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/twitchtv/twirp-example/rpc/helloworld"
)

func main() {
	client := pb.NewHelloWorldProtobufClient("http://localhost:8080", &http.Client{})

	resp, err := client.Hello(context.Background(), &pb.HelloReq{Subject: "World"})
	if err == nil {
		fmt.Println(resp.Text) // prints "Hello World"
	}
}
```

## Why this is good

There's no need to worry about JSON serialization or HTTP verbs/routes! Twirp
[routing and serialization](routing.md) handles that for you, reducing the risk
of introducing bugs. Both JSON and Protobuf are supported. The
[Protobuf protocol](https://developers.google.com/protocol-buffers/docs/proto3)
is designed to allow backwards compatible changes (unlike JSON, it is trivial to
rename fields), it is faster than JSON and also works well as documentation for
your service, because it is easy to tell what a service does by just looking at
the Proto file.

## What's next?

 * [Install Twirp](install.md): instructions to install or upgrade Twirp tools
   for code auto-generation (protoc, protoc-gen-go and protoc-gen-twirp).
 * [Usage Example](example.md): step by step guide to build an awesome
   Haberdasher service.
 * [How Twirp routes requests](routing.md): learn more about how Twirp works
   under the covers.
