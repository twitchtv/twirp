A major release that includes a protocol spec update. The jump from v5 to v7 is to avoid confusion with issues and comments that were targeting v6 and were never released (see [achieved protocol spec v6](https://twitchtv.github.io/twirp/docs/spec_v6.html)).

Changes in the Protocol spec v7:

 * Twirp routes can have any prefix: `<prefix>/<package>.<Service>/<Method>` section. In v5 the prefix was mandatory was always `/twirp`.
 * Error code `ResourceExhausted` maps to HTTP code `429`. In v5 it mapped to `403`.

Changes included in the Go Twirp v7 release:

 * #264 Optional Twirp Prefix. Implements the proposal #263 with the optional prefix, using an option to specify a different prefix than the default value "/twirp". The default value ensures backwards compatibility when updating the service. Using different prefixes, it is now possible to mount the same Twirp service in multiple routes, which may help migrating existing services from the "/twirp" prefix to a new one.
 * #264 also introduces [server options on the server constructor](https://github.com/twitchtv/twirp/pull/264#issuecomment-686170407). Adding server hooks should now done through a server option. The server options are available in the new version of the `twirp` package, servers generated with the new version require the `twirp` package to be updated as well.
 * #270 ResourceExhausted error code to HTTP 429 status code. To implement the new spec for this error, so it is clear that this error code can be used for rate limiting. This change may affect middleware if it depends on a specific status code being used.
 * #271 Server JSON responses using EmitDefaults: to eliminate a common pain point when debugging a Twirp service with cURL for the first time, now the default JSON serialization will include all the fields in the response, even if they are the zero-values. This can be reverted to the previous behavior (skip zero-values) with the server option: `twirp.WithServerJSONSkipDefaults(true)`.
 * #257 Server handles routes with non-CamelCased service and method names. Fixes some points from the issue #244, that happens when Twirp services from different languages (e.g. Go and Python) have a proto definition that is not using the recommended [Protobuf Style Guide](https://developers.google.com/protocol-buffers/docs/style#services). Now Go services can handle literal routes and properly implements the routing spec.
 * #268 Bufgfix: Allow to update context from server Error hooks, so they can communicate with the ResponseSent hook if needed.
