---
id: migrate_to_twirp
title: Migrate APIs to Twirp
sidebar_label: Migrate APIs to Twirp
---

## Migrate a REST API to Twirp

Migrating existing REST/JSON APIs to Twirp is a case-by-case scenario. There's no one-size-fits-all process for it.

REST/JSON APIs have a huge variety of ways to be defined, but Twirp APIs are restricted to requests in the form `POST [<prefix>]/[<package>.]<Service>/<Method>` (see [Routing and Serialization](routing.md)). That limitation is what allows to build simple and consistent API schemas, but it also means that each REST API endpoint will have a different way to be translated into a Twirp endpoint.

The good news is that you can have Twirp services working together with regular REST APIs, and slowly migrate your older APIs as needed. The general approach involves the following steps:

1.  Identify all the API endpoints in your old service, including request parameters, valid responses and possible errors. Also, identify all the API callers (upstream services) and let them know about the migration, they will need to update their client to use the new Twirp client when available.
2.  Write a Protobuf schema with the same endpoints. Remember that Twirp endpoints depend on the Protobuf message used as request, but can also read [HTTP headers](headers.md) if needed. Respect Protobuf [naming and best practices](best_practices.md) when possible, but don't try and significantly change the API design of your application during the migration, it will be a lot easier to migrate callers to the new API if the new endpoints closely resemble the old endpoints. Try to use parameter types that are similar to the older parameters. Use comments on the schema to clarify edge cases.
3.  If the API is too big, it may be better to only migrate a few endpoints first. Twirp APIs are good at evolving over time, you can always add more endpoints later.
4.  Generate code and implement your new API methods. You may be able to re-use some of the older code from the previous endpoints. One way to do this is to implement the new service in a subfolder inside the same project, and then mount the new Twirp handler in a sub-route (you can use the default `/twirp` prefix, or specify a different prefix path if needed).
5.  Make sure to track stats about API endpoint usage, so you can see the traffic being migrated to the new endpoints. Using [hooks or interceptors](hooks.md), or HTTP [middleware](mux.md) can be useful.
6.  The API callers (upstream services) need to import the new Twirp client and update their calls to use the new Twirp API. Here is where making the new API similar to the old one will be the most useful. Some services may be able to just replace the old client with the new one and start sending traffic to the new endpoints. If they have a decent staging/canary environment and a good deploy system that allows for quick rollbacks, that could be enough to test that the new API is handling the new traffic properly. In other cases, you will need to implement a slider to migrate traffic slowly (10%, 50%, 100%). The slider can be implemented in different ways, either with if statements, e.g. `if r.Intn(100) < 10 { callNewClient() } else { callOldClient() }`, or in more complex cases, with a custom client that implements the same Twirp service interface, but it is able to split the traffic between the old API and the new API.
7.  Cleanup the old code when all the traffic is migrated to the new service. If the old API is impossible to remove completely, it may be useful to re-implement each endpoint as a shim that calls the new Twirp endpoint in the server side.
