---
id: migrate_to_twirp
title: Migrate APIs to Twirp
sidebar_label: Migrate APIs to Twirp
---

## Migrate REST/JSON APIs to Twirp

Migrating existing REST/JSON APIs to Twirp is a case-by-case scenario. Unfortunately, it involves creating a new service, generating a new client, and migrating callers to use the new endpoints in the new client.

Twirp APIs are restricted to requests in the form `POST [<prefix>]/[<package>.]<Service>/<Method>` (see [Routing and Serialization](routing.md)). That limitation is what allows building simple and consistent API schemas. In the other hand, REST/JSON APIs usually don't have strongly defined schemas and tend to have edge cases that need special attention when being translated into Twirp endpoints.

The process is a case-by-case scenario, but we can highlight a few common steps:

1.  Identify all the API endpoints in your old service, including request parameters, valid responses and possible errors. Also, identify all the API callers (upstream services) and let them know about the migration, they will soon need to update their client to use the new Twirp client when available.
2.  Write a Protobuf schema. Respect [Protobuf naming and best practices](https://developers.google.com/protocol-buffers/docs/style) when possible, but don't try and significantly change the API design of your application during the migration; it will be a lot easier to migrate callers to the new API if the new endpoints are closely related to the old endpoints. Try to use parameter types that are similar to the old parameters. Use comments on the schema to clarify edge cases.
3.  If the API is too big, it may be better to only migrate a few endpoints first. Protobuf schemas are good at evolving over time, you can always complete the migration for a few endpoints and add more endpoints later.
4.  Generate code and implement your new API methods. You may be able to re-use some of the older code from the previous endpoints. One way to do this is to implement the new service in a subfolder inside the same project, and then mount the new Twirp handler in a sub-route (see [Muxing Twirp with other services](mux.md)).
5.  Make sure to track stats about API endpoint usage, so you can see the traffic being migrated to the new endpoints. Use [hooks, interceptors](hooks.md) and/or HTTP [middleware](mux.md) as needed.
6.  When the new Twirp service is ready, the API callers (upstream services) need to import the new Twirp client and update their calls to use the new Twirp API. This is why making the new API similar to the old one is useful. Some services may be able to just replace the old client with the new one and start sending traffic to the new endpoints, relying on their staging/canary environment for testing. Other services may want to implement a rollout slider to migrate traffic slowly (10%, 50%, etc.). The slider can be implemented in different ways. Here's a very naive implememtation in Go to allow 10% traffic on the new client: `if r.Intn(100) < 10 { callNewClient() } else { callOldClient() }`. In more complex cases, it is also possible to make client that implements the same Twirp service interface, but with options to split the traffic between the new API and the old API.
7.  When all traffic is migrated, don't forget to clean up the old code. If the old API is impossible to remove completely, it may be useful to re-implement old endpoints as shim calls to the new Twirp endpoints.
