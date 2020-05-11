# Contributing

Thanks for helping make Twirp better! This is great!

## Twirp Principles

Your contribution will go more smoothly if it aligns well with the project priorities. Twirp has been developed by twitch.tv and it has been in production for years. As many internal systems, teams and processes depend on Twirp, keeping backwards compatibility is taken very seriously.

Design principles:

 * Simplicity. Keep serialization and routing simple and intuitive.
 * Small api and interface. Reduce future support burden.
 * Prevent user errors. Avoid surprising behavior.
 * Pragmatism over being bleeding edge.
 * Hooks are for observability instead of control flow, so it is easier to debug Twirp systems.
 * No flags/options for code generation, so generated code is predictable across versions and platforms.
 * As few dependencies as possible, so it is easier to integrate and upgrade.
 * Prefer generated code over shared libraries between services and clients, so it is easier to implement changes without breaking compatibility.

Contributions that are welcome:

 * Security updates.
 * Performance improvements.
 * Supporting new versions of Golang and Protobug.
 * Documentation.
 * Making Twirp easier to integrate with other useful tools.

In the other hand, contributions that contradict the following priorities will be more difficult:

 * Backwards compatibility is very important. Changes that break compatibility will see resistance, even when they only break a small use case. There must be a good reason behind a major version update.
 * Complex and involved features should be avoided. Twirp was designed to be easy to use and easy to maintain. See the streaming proposal as an example of a complex feature that was not able to make it into a major release: https://github.com/twitchtv/twirp/issues/3.
 * Features that could be implemented in a separate library should go in a separate library. Twirp allows easy integrtion with 3rd party code.


## Report an Issue

If you have run into a bug or want to discuss a new feature, please [file an issue](https://github.com/twitchtv/twirp/issues). If you'd rather not publicly discuss the issue, please email security@twitch.tv.

## Contributing Code with Pull Requests

Twirp uses github pull requests. Fork, hack away at your changes and submit. Most pull requests will go through a few iterations before they get merged. Different contributors will sometimes have different opinions, and often patches will need to be revised before they can get merged.

### Requirements

 * Twirp officially supports the last 3 releases of Go.
 * The Python implementation uses Python 2.7.
 * Protoc v3 to generate code.
 * For linters and other tools, we use [retool](https://github.com/twitchtv/retool). If `make setup` is not able to install it, you can install it in your path with `go get github.com/twitchtv/retool` and then install tools with `retool build`.

### Running tests

Generally you want to make changes and run `make`, which will install all
dependencies we know about, build the core, and run all of the tests that we
have against Go and Python code. A few notes:

 * Make sure to clone the repo on `$GOPATH/src/github.com/twitchtv/twirp`
 * Run Go unit tests with `make test_core`, or just the tests with `go test -race ./...`.
 * Most tests of the Go server are in `internal/twirptest/service_test.go`.
 * Integration tests running the full stack in both Golang and Python auto-generated clients are in the [clientcompat](./clientcompat) directory.

## Contributing Documentation

Twirp's docs are generated with [Docusaurus](https://docusaurus.io/). You can
safely edit anything inside the [docs](./docs) directory, adding new pages or
editing them. You can edit the sidebar by editing
[website/sidebars.json](./website/sidebars.json).

Then, to render your changes, run docusaurus's local server. To do this:

 1. [Install docusaurus on your machine](https://docusaurus.io/docs/en/installation.html).
 2. `cd website`
 3. `npm start`
 4. Navigate to http://localhost:3000/.

## Making a New Release

Releasing versions is the responsibility of the core maintainers. Most people
can skip this section.

Twirp uses Github releases. To make a new release:

 1. Merge all changes that should be included in the release into the master
    branch.
 2. Update the version constant in `internal/gen/version.go`. Make sure to respect [semantic versioning](http://semver.org/): `v<major>.<minor>.<patch>`.
 3. Add a new commit to master with a message like "Version vX.X.X release".
 4. Tag the commit you just made: `git tag <version number>` and `git push
    origin --tags`
 5. Run `make release_gen` to generate release assets in the `release`
    directory. This requires Docker to be installed.
 6. Go to Github https://github.com/twitchtv/twirp/releases and
    "Draft a new release".
 7. Make sure to document changes, specially when upgrade instructions are
    needed.
 8. Upload all files in the `release` directory as part of the release.


## Code of Conduct

This project has adopted the [Amazon Open Source Code of Conduct](https://aws.github.io/code-of-conduct).
For more information see the [Code of Conduct FAQ](https://aws.github.io/code-of-conduct-faq) or contact
opensource-codeofconduct@amazon.com with any additional questions or comments.


## Licensing

See the [LICENSE](https://github.com/twitchtv/twirp/blob/master/LICENSE) file for our project's licensing. We will ask you to confirm the licensing of your contribution.

We may ask you to sign a [Contributor License Agreement (CLA)](http://en.wikipedia.org/wiki/Contributor_License_Agreement) for larger changes.
