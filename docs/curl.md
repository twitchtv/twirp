---
id: "curl"
title: "cURL"
sidebar_label: "cURL"
---

You can access a Twirp service with cURL, using either JSON or Protobuf.

## Example

With the following HelloWorld service defined in this proto file:

```proto
syntax = "proto3";
package example.helloworld;

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

Assuming a service generated from this definition is running in `http://localhost:8080` with the default
"/twirp" prefix, you can call it with cURL by following the routing rules.

### JSON

Use the header `Content-Type: application/json` to signal that the request and response are JSON:

```sh
curl --request "POST" \
    --header "Content-Type: application/json" \
    --data '{"subject": "World"}' \
    http://localhost:8080/twirp/example.helloworld.HelloWorld/Hello
```

The service should respond with something like this:

```json
{"text": "Hello World"}
```

NOTE: Twirp uses [proto3-json mapping](https://developers.google.com/protocol-buffers/docs/proto3#json),
which means that empty fields are excluded. If you specify an empty request `--data '{}'` it will be
interpreted as zero-values. Zero-values are also excluded on responses. In this example,
if the service responded with an empty "text" field, the response you will see is empty `{}`.

### Protobuf

Use the header `Content-Type: application/protobuf` to signal that the request and response are Protobuf.
Use the `protoc` tool to encode and decode the Protobuf messages into readable key-values:

```sh
echo 'subject:"World"' \
	| protoc --encode example.helloworld.HelloReq ./rpc/helloworld/service.proto \
	| curl -s --request POST \
      --header "Content-Type: application/protobuf" \
      --data-binary @- \
      http://localhost:8080/twirp/example.helloworld.HelloWorld/Hello \
	| protoc --decode example.helloworld.HelloResp ./rpc/haberdasher/service.proto
```

The service should respond with something like this:

```
text:"Hello World"
```

### Errors

Twirp error responses are always JSON, even if the request is done in Protobuf. A Twirp error response looks like this:

```json
{"code": "internal", "msg": "Something went wrong"}
```
