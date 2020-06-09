---
id: "spec_v6"
title: "Twirp Wire Protocol"
sidebar_label: "Version 6 (archived)"
---

This is a historical document for the Twirp wire protocol "v6", that was never released. At the time of writing, all existing production systems are using the "v5" protocol.

## Why is this specific draft of the v6 protocol not going to be released?

The proposal for the v6 protocol has been standing without progress for too long. The wire protocol is the glue connecting all the systems in production, that are using different implementations across multiple languages. Because of this, it is very difficult to plan a path forward without having a negative impact on the Twirp ecosystem.

## Feature Reference

The v6 protocol introduced a change to the URL schema, and added support for streams.

### Streaming API

Issue: https://github.com/twitchtv/twirp/issues/70

Handling streams out of the box. Not implemented because streams require assumptions about how the connection state is managed, they are complex and not required by the majority of Twirp users. Websockets or gRPC may be good alternatives.

There may still be reasons to implement streams in the future. If that happens, it will be managed separately from this v6 protocol draft.


### HTTP Routes without the /twirp prefix

Issue: https://github.com/twitchtv/twirp/issues/55

Twirp routes have a `/twirp` prefix that is inconvenient in some cases. There was a proposal to allow different route prefixes, but too much time has passed since the proposal, and too many services depend on this prefix. Instead of updating the Twirp protocol, issue comments offer how to manage this issue from middleware.
