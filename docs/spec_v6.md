---
id: "spec_v6"
title: "Twirp Wire Protocol (v6)"
sidebar_label: "Version 6 (Archived)"
---

This is a historical document for the Twirp wire protocol "v6", that was never released. At the time of writing, all existing production systems were using the "v5" protocol. New systems will implement "v7" and above.

## Why was v6 never released?

The proposal for the v6 protocol has been standing without progress for too long. The streaming proposal was not simple enough to be officially released. The wire protocol is the glue connecting all the systems in production, once formalized it needs to be implemented by all different implementations across multiple languages. Because of this, it is very difficult to plan a path forward without having a negative impact on the Twirp ecosystem.

### Feature: Streaming API

Issue: https://github.com/twitchtv/twirp/issues/70

Handling streams out of the box. Not implemented because streams require assumptions about how the connection state is managed, they are complex and not required by the majority of Twirp users. Websockets or gRPC may be good alternatives.

There may still be reasons to implement streams in the future. If that happens, it will be managed separately from this v6 protocol draft.


### Feature: HTTP Routes update

Issue: https://github.com/twitchtv/twirp/issues/55

Twirp routes have a `/twirp` prefix that is inconvenient in some cases. The different proposals to allow different routes were not entirely backwards compatible, and they were waiting for the Streaming API branch to be finalized. Too much time has passed since those routing proposals, and the backwards compatibility requirements have become more important.

The "v7" protocol introduces an optional prefix for routing (where `/twirp` is the default prefix): https://github.com/twitchtv/twirp/pull/264

