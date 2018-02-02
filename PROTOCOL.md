# Twirp Wire Protocol

This document defines the Twirp wire protocol over HTTP. The
current protocol version is v5.

## Overview

The Twirp wire protocol is a simple RPC protocol based on HTTP and
Protocol Buffers (proto). The protocol uses HTTP URLs to specify the
RPC endpoints, and sends/receives proto messages as HTTP
request/response bodies.

To use Twirp, developers first define their APIs using proto files,
then use Twirp tools to generate the client and the server libraries.
The generated libraries implement the Twirp wire protocol, using the
standard HTTP library provided by the programming language runtime or
the operating system. Once the client and the server are implemented,
the client can communicate with the server by making RPC calls.

The Twirp wire protocol supports both binary and JSON encodings of
proto messages, and works with any HTTP client and any HTTP version.

### URLs

In [ABNF syntax](https://tools.ietf.org/html/rfc5234), Twirp's URLs
have the following format:

**URL ::= Base-URL "/twirp/" [ Package "." ] Service "/" Method**

The Twirp wire protocol uses HTTP URLs to specify the RPC
endpoints on the server for sending the requests. Such direct mapping
makes the request routing simple and efficient. The Twirp URLs have
the following components.

* **Base-URL** is the virtual location of a Twirp API server, which is
  typically published via API documentation or service discovery.
  Currently, it should only contain URL `scheme` and `authority`. For
  example, "https://example.com".

* **Package** is the proto `package` name for an API, which is often
  considered as an API version. For example,
  `example.calendar.v1`. This component is omitted if the API
  definition doesn't have a package name.

* **Service** is the proto `service` name for an API. For example,
  `CalendarService`.

* **Method** is the proto `rpc` name for an API method. For example,
  `CreateEvent`.

### Requests

Twirp always uses HTTP POST method to send requests, because it
closely matches the semantics of RPC methods.

The **Request-Headers** are normal HTTP headers. The Twirp wire
protocol uses the following headers.

* **Content-Type** header indicates the proto message encoding, which
  should be one of "application/protobuf", "application/json". The
  server uses this value to decide how to parse the request body,
  and encode the response body.

The **Request-Body** is the encoded request message, contained in the
HTTP request body. The encoding is specified by the `Content-Type`
header.

### Responses

The **Response-Headers** are just normal HTTP response headers. The
Twirp wire protocol uses the following headers.

* **Content-Type** The value should be either "application/protobuf"
  or "application/json" to indicate the encoding of the response
  message. It must match the "Content-Type" header in the request.

The **Request-Body** is the encoded response message contained in the
HTTP response body. The encoding is specified by the `Content-Type`
header.

### Example

The following example shows a simple Echo API definition and its
corresponding wire payloads.

The example assumes the server base URL is "https://example.com".

```proto
syntax = "proto3";

package example.echoer;

service Echo {
  rpc Hello(HelloRequest) returns (HelloResponse);
}

message HelloRequest {
  string message;
}

message HelloResponse {
  string message;
}
```

**Proto Request**

```
POST /twirp/example.echoer.Echo/Hello HTTP/1.1
Host: example.com
Content-Type: application/protobuf
Content-Length: 15

<encoded HelloRequest>
```

**JSON Request**

```
POST /twirp/example.echoer.Echo/Hello HTTP/1.1
Host: example.com
Content-Type: application/json
Content-Length: 27

{"message":"Hello, World!"}
```

**Proto Response**

```
HTTP/1.1 200 OK
Content-Type: application/protobuf
Content-Length: 15

<encoded HelloResponse>
```

**JSON Response**

```
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 27

{"message":"Hello, World!"}
```

## Errors

Twirp error responses are always JSON-encoded, regardless of
the request's Content-Type, with a corresponding
`Content-Type: application/json` header. This ensures that
the errors are human-readable in any setting.

Twirp errors are a JSON object with three keys:

* **code**: One of the Twirp error codes as a string.
* **msg**: A human-readable message describing the error
  as a string.
* **meta**: An object with string keys and values holding
  arbitrary additional metadata describing the error.

Example:
```
{
    "code": "permission_denied",
    "msg": "thou shall not pass",
    "meta": {
        "target": "Balrog"
    }
}
```

For more information, see https://github.com/twitchtv/twirp/wiki/Errors.
