# clientcompat #

clientcompat is a tool to test the compatibility of Twirp client
implementations.

## Usage ##
`clientcompat -client=<binary>`

The client should be generated for the compattest.proto file found in this
directory.

The client binary must accept, over stdin, a protobuf-encoded
ClientCompatMessage (defined in clientcompat.proto). This message contains a
`service_address`, which is the address of a reference-implementation
CompatService to talk to, a `method`, which specifies which RPC of the
CompatService should be called, and an embedded, proto-encoded `request` which
the client should send.

If the server sends an error, then the client should parse the error and write
the error code string ("internal", "unauthenticated", etc) to stderr. If the
server doesn't send an error, the client should encode the response message it
received as protobuf and write it to stdout.

## Example ##

The [gocompat](./gocompat) subdirectory contains an example implementation which
can be run with clientcompat to prove that the Go client implementation works.
