---
id: "proto_and_json"
title: "Twirp's Serialization Schemes"
sidebar_label: "Protobuf and JSON"
---

Twirp can handle both Protobuf and JSON encoding for requests and responses.

This is transparent to your service implementation. Twirp parses the HTTP
request (returning an Internal error if the `Content-Type` or the body are
invalid) and converts it into the request struct defined in the interface. Your
implementation returns a response struct, which Twirp serializes back to a
Protobuf or JSON response (depending on the request `Content-Type`).

See [the spec](spec.md) for more details on routing and serialization.

Twirp can generates two types of clients for your service:

 * `New{{Service}}ProtobufClient`: makes Protobuf requests to your service.
 * `New{{Service}}JSONClient`: makes JSON requests to your service.

### Which one should I use, ProtobufClient or JSONClient?

You should use the **ProtobufClient**.

Protobuf uses fewer bytes to encodethan JSON (it is more compact), and it
serializes faster.

In addition, Protobuf is designed to gracefully handle schema
updates. Did you notice the numbers added after each field? They allow you to
change a field and it still works if the client and the server have a different
versions (that doesn't work with JSON clients).

### If Protobuf is better, why does Twirp support JSON?

You will probably never need to use a Twirp **JSONClient** in Golang, but having
your servers automatically handle JSON requests is still very convenient. It
makes it easier to debug (see
[cURL requests](https://github.com/twitchtv/twirp/wiki/HTTP-Routing-and-Serialization#making-requests-on-the-command-line-with-curl)),
allows to easily write clients in other languages like Python, or make REST
mappings to Twirp services.

The JSON client is generated to provide a reference for implementations in other
languages, and because in some rare circumstances, binary encoding of request
bodies is unacceptable, and you just need to use JSON.
