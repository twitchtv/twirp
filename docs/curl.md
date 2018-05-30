---
id: "curl"
title: "cURL"
sidebar_label: "cURL"
---

Twirp allows you to cURL your service with either Protobuf or JSON.

## Example

A cURL request to the Haberdasher example `MakeHat` RPC with the following request and reply could be executed as Protobuf or JSON with the snippets further below.:

Request proto:
```
message Size {
   int32 inches = 1;
}
```

Reply proto:

```
message Hat {
  int32 inches = 1;
  string color = 2;
  string name = 3;
}
```

### Protobuf

```sh
echo "inches:10" \
	| protoc --proto_path=$GOPATH/src --encode twirp.example.haberdasher.Size ./rpc/haberdasher/service.proto \
	| curl -s --request POST \
                  --header "Content-Type: application/protobuf" \
                  --data-binary @-
                  http://localhost:8080/twirp/twirp.example.haberdasher.Haberdasher/MakeHat \
	| protoc --proto_path=$GOPATH/src --decode twirp.example.haberdasher.Hat ./rpc/haberdasher/service.proto
```

We signal Twirp that we're sending Protobuf data by setting the `Content-Type` as `application/protobuf`.

The request is:

```
inches:10
```

The reply from Twirp would look something like this:

```
inches:1
color:"black"
name:"bowler"
```

### JSON

```sh
curl --request "POST" \
     --location "http://localhost:8080/twirp/twirp.example.haberdasher.Haberdasher/MakeHat" \
     --header "Content-Type:application/json" \
     --data '{"inches": 10}' \
     --verbose
```

We signal Twirp that we're sending JSON data by setting the `Content-Type` as `application/json`.

The request is:

```json
{"inches": 10}
```

The JSON response from Twirp would look something like this:

```json
{"inches":1, "color":"black", "name":"bowler"}
```
