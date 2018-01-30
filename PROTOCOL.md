# Twirp Wire Protocol

This document defines the Twirp wire protocol over HTTP.

## Conventions

The requirement level keywords "MUST", "MUST NOT", "REQUIRED",
"SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY",
and "OPTIONAL" used in this document are to be interpreted as
described in [RFC 2119](https://www.ietf.org/rfc/rfc2119.txt).

The grammar rules used in this document are using [ABNF
syntax](https://tools.ietf.org/html/rfc5234).

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
However, certain capabilities may be limited by the actual HTTP
library being used.

### URLs

**URL ::= Base-URL "/" [ Package "." ] Interface "/" Method**

The Twirp wire protocol uses HTTP URLs to directly specify the RPC
endpoints on the server for sending the requests. such direct mapping
makes the request routing simple and efficient. The Twirp URLs have
the following components.

* **Base-URL** is the virtual location of a Twirp API server, which is
  typically published via API documentation or service discovery.  For
  example, "https://example.com/apis".

* **Package** is the proto `package` name for an API, which is often
  considered as an API version. For example,
  `example.calendar.v1`. This component is omitted if the API
  definition doesn't have a package name.

* **Interface** is the proto `service` name for an API. For example,
  `CalendarService`.

* **Method** is the proto `rpc` name for an API method. For example,
  `CreateEvent`.

### Requests

**Request ::= Request-Headers Request-Body**

Twirp always uses HTTP POST method to send requests, because it
closely matches the semantics of RPC methods.

The **Request-Headers** are normal HTTP headers. The Twirp wire
protocol uses the following headers.

* **Authorization** header is often used to pass user credentials
  from the client to the server, such as OAuth access token or
  JWT token.

* **Content-Type** header indicates the proto message encoding, which
  should be one of "application/x-protobuf", "application/json". The
  server uses this value to decide how to parse the request body,
  and encode the response body.

* **User-Agent** header indicates the client application and its
  runtime environment. While the server should not use this
  information for request processing, this header is heavily used
  for analytics and troubleshooting purposes.

* **RPC-Timeout** header indicates the client-specified request
  timeout in seconds, such as "10". If **RPC-Timeout** is omitted, the
  server should use a pre-configured timeout value, by default it
  should be 10 seconds.

The **Request-Body** is the encoded request message, contained in the
HTTP request body. The encoding is specified by the `Content-Type`
header.

### Responses

**Response ::= Response-Headers Response-Body**

The **Response-Headers** are just normal HTTP response headers. The
Twirp wire protocol uses the following headers.

* **Content-Type** The value should be either "application/x-protobuf"
  or "application/json" to indicate the encoding of the response
  message. It must match the "Content-Type" header in the request.

The **Request-Body** is the encoded response message contained in the
HTTP response body. The encoding is specified by the `Content-Type`
header.

## Example

The following example shows a simple Echo API definition and its
corresponding wire payloads.

The example assumes the server base URL is "https://example.com".

```proto
package twirp;

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
POST /twirp.Echo/Hello HTTP/1.1
Host: example.com
Content-Type: application/x-protobuf
Content-Length: 15

<encoded HelloRequest>
```

**JSON Request**

```
POST /twirp.Echo/Hello HTTP/1.1
Host: example.com
Content-Type: application/json
Content-Length: 27

{"message":"Hello, World!"}
```

**Proto Response**

```
HTTP/1.1 200 OK
Content-Type: application/x-protobuf
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

If an error occurs when the server processes a request, the server
must return an error payload as the response message, and correctly
set the HTTP status code. Please see
[`google.rpc.Code`](https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto)
on how to map typical server errors to HTTP status codes.

### Timeout errors

For a single request, there is a client-specified timeout and a
server-configured timeout. If a request misses the server-configured
timeout, the server must return a `503` error. If a request misses
the client-specified timeout that it is shorter then the server-
configured timeout, the server must return a `504` error.
This allows more accurate measurement of the server availability.

### Network errors

If a client fails to reach the server due to network errors, the
client library must report HTTP status code `502` to the client
application. This helps users distinguishing network errors from
server errors.
