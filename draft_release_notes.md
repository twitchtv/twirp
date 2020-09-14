A major release that includes a protocol spec update. The jump from v5 to v7 is to avoid confusion with issues and comments that were targeting v6 and were never released (see [archived protocol spec v6](https://twitchtv.github.io/twirp/docs/spec_v6.html)).

Changes in the Protocol spec v7:

 * Twirp routes can have any prefix: `<prefix>/<package>.<Service>/<Method>`. In v5 the prefix was mandatory could only be `/twirp`.
 * Error code `ResourceExhausted` maps to HTTP code `429`. In v5 it mapped to `403`.

Changes included in the Go Twirp v7 release:

 * #264 Optional Twirp Prefix. Implements the proposal #263 with the optional prefix, using an option to specify a different prefix than the default value "/twirp". The default value ensures backwards compatibility when updating the service. Using different prefixes, it is now possible to mount the same Twirp service in multiple routes, which may help migrating existing services from the "/twirp" prefix to a new one.
 * #264 also introduces [server options on the server constructor](https://github.com/twitchtv/twirp/pull/264#issuecomment-686170407). Adding server hooks should now done through a server option. The server options are available in the new version of the `twirp` package, servers generated with the new version require the `twirp` package to be updated as well.
 * #270 ResourceExhausted error code changed to 429 HTTP status code. This may affect middleware if it depends on a specific status code being used.
 * #271 Server JSON responses using EmitDefaults: responses include all fields, even if they have zero-values. This can be reverted to the previous behavior (skip zero-values) with the server option: `twirp.WithServerJSONSkipDefaults(true)`.
 * #257 Go services handle literal routes. Fixing part of the issue #244, affecting cross-language communications when the proto definitions are not using the recommended [Protobuf Style Guide](https://developers.google.com/protocol-buffers/docs/style#services).
 * #268 Bufgfix: Allow to update context from server Error hooks, so they can communicate with the ResponseSent hook if needed.
