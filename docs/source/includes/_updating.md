# Updating Twirp

> If you are using [retool](https://github.com/twitchtv/retool), update using the following command:

```bash
$ retool upgrade github.com/twitchtv/twirp/protoc-gen-twirp v5.0.0
```

Twirp releases are tagged with semantic versioning and releases are managed by Github. See the [releases](https://github.com/twitchtv/twirp/releases) page.

To stay up to date, you update protoc-gen-twirp and regenerate your code. It's simple with [retool](https://github.com/twitchtv/retool).

<div class="clear"></div>

> Use `go install` to update twirp if you don't have retool:

```golang
$ cd $GOPATH/src/github.com/twitchtv/twirp
$ git checkout v5.0.0
$ go install ./protoc-gen-twirp
```

If you're not using retool, [you should be using retool](https://github.com/twitchtv/retool)! But you can also do a system-wide install with checking out the package new version and using go install:

<div class="clear"></div>

With the new version of protoc-gen-twirp, you can re-generate code to update your servers. Then, any of the clients of your service can update their vendored copy of your service to get the latest version.
