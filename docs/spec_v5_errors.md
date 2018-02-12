---
id: "spec_v5_errors"
title: "Twirp Errors Wire Protocol"
sidebar_label: "Version 5 (Latest)"
---

This document defines the Twirp wire protocol over HTTP for Errors.
For more information on the wire protocol see [spec v5](spec_v5.md).

## Twirp Errors

Erros are non-200 JSON responses with the keys:

* **code**: (string) One of the valid Twirp error codes.
* **msg**: (string) Human-readable message describing the error.
* **meta**: (optional object) arbitrary additional metadata 
  describing the error. Keys and values must be strings.

Example:
```json
{
  "code": "internal",
  "msg": "Something went wrong"
}
```

### Content-Type JSON

Twirp error responses are always JSON-encoded, the response Content-Type
header is always `application/json`, regardless of the request's 
Content-Type (protbuf requests also respond with json errors).
This ensures that errors are human-readable and easy to parse by intermediary proxies.

Example Error Response:
```
HTTP/1.1 500 Internal Server Error
Content-Type: application/json
Content-Length: 48

{"code":"internal","msg":"Something went wrong"}
```

### Twirp Error Codes

The possible values for the Error "code" are intentionaly similar to
[gRPC status codes](https://godoc.org/google.golang.org/grpc/codes):
	
* **canceled**: The operation was cancelled (typically by the caller).
* **unknown**: Unknown error. For example when handling errors raised by APIs 
  that do not return enough error information.
* **invalid_argument**: the client specified an invalid argument. It
  indicates arguments that are problematic regardless of the state of the
  system (i.e. a malformed file name, required argument, number out of range, etc.).
* **deadline_exceeded**: Operation expired before completion. For operations
  that change the state of the system, this error may be returned even if the
  operation has completed successfully (timeout).
* **not_found**: Some requested entity was not found.
* **bad_route**: The requested URL path wasn't routable to a Twirp
  service and method. This is returned by the generated server, and usually
  shouldn't be returned by applications. Instead, applications should use
  "not_found" or "unimplemented".
* **already_exists**: An attempt to create an entity failed because one
  already exists.
* **permission_denied**: The caller does not have permission to execute
  the specified operation. It must not be used if the caller cannot be
  identified (use "unauthenticated" instead).
* **unauthenticated**: The request does not have valid authentication
  credentials for the operation.
* **resource_exhausted**: Some resource has been exhausted, perhaps a
  per-user quota, or perhaps the entire file system is out of space.
* **failed_precondition**: The operation was rejected because the system is
  not in a state required for the operation's execution. For example, doing
  an rmdir operation on a directory that is non-empty, or on a non-directory
  object, or when having conflicting read-modify-write on the same resource.
* **aborted**: The operation was aborted, typically due to a concurrency
  issue like sequencer check failures, transaction aborts, etc.
* **out_of_range**: The operation was attempted past the valid range. 
  For example, seeking or reading past end of a paginated collection.
  Unlike "invalid_argument", this error indicates a problem that may be fixed if
  the system state changes (i.e. adding more items to the collection).
  There is a fair bit of overlap between "failed_precondition" and "out_of_range".
  We recommend using "out_of_range" (the more specific error) when it applies so
  that callers who are iterating through a space can easily look for an
  "out_of_range" error to detect when they are done.
* **unimplemented**: The operation is not implemented or not
  supported/enabled in this service.
* **internal**: When some invariants expected by the underlying system
  have been broken. In other words, something bad happened in the library or
  backend service. Twirp specific issues like wire and serialization problems
  are also reported as "internal" errors.
* **unavailable**: The service is currently unavailable. This is most
  likely a transient condition and may be corrected by retrying with a
  backoff.
* **data_loss**: Unrecoverable data loss or corruption.


### HTTP status code

Twirp services must respond with status `200` for non-error responses.
For error responses, the staus depends on the Twirp Error Code:

```
Twirp Error Code        HTTP status code
---------------------- -----------------------
"canceled"            | 408 RequestTimeout
"unknown"             | 500 Internal Server Error
"invalid_argument"    | 400 BadRequest
"deadline_exceeded"   | 408 RequestTimeout
"not_found"           | 404 Not Found
"bad_route"           | 404 Not Found
"already_exists"      | 409 Conflict
"permission_denied"   | 403 Forbidden
"unauthenticated"     | 401 Unauthorized
"resource_exhausted"  | 403 Forbidden
"failed_precondition" | 412 Precondition Failed
"aborted"             | 409 Conflict
"out_of_range"        | 400 Bad Request
"unimplemented"       | 501 Not Implemented
"internal"            | 500 Internal Server Error
"unavailable"         | 503 Service Unavailable
"dataloss"            | 500 Internal Server Error
```

### Non-Twirp Errors from Intermediary Proxies

It is also possible for Twirp Clients to receive non-200 responses with invalid 
Twirp Errors (non JSON, or with invalid keys), this responses may be
generated by proxy middleware or network issues. Twirp Clients should build a
valid Twirp error depending on the HTTP status of those invalid responses:

```
HTTP status code           Twirp Error Code
------------------------- -----------------------
3xx (redirects)          | "internal"
400 Bad Request          | "internal"
401 Unauthorized         | "unauthenticated"
403 Forbidden            | "permission_denied"
404 Not Found            | "bad_route"
429 Too Many Requests    | "unavailable"
502 Bad Gateway          | "unavailable"
503 Service Unavailable  | "unavailable"
504 Gateway Timeout      | "unavailable"
... other                | "unknown"
```

Additional metadata should be added to make it easy to identify intermediary errors:
* `"http_error_from_intermediary": "true"`
* `"status_code": string` (original status code on the HTTP response, e.g. `"500"`).
* `"body": string` (original non-Twirp error response as string).
* `"location": url-string` (only on 3xx reponses, matching the `Location` header).


### Metadata

In adition to `code` and `msg`, Twirp Error responses may have a `meta` field with
arbitrary string metadata. For example:
```json
{
	"code": "invalid_argument",
	"msg": "please use a smaller size",
	"meta": {
		"argument": "size",
		"max_allowed": "1000"
	}
}
```

Error metadata can only have string values. This is to simplify error parsing by clients.
If your service requires errors with complex metadata, you should consider adding client
wrappers on top of the auto-generated clients, or just include business-logic errors as
part of the Protobuf messages (add an error field to proto messages).


