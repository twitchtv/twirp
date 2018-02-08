# Contributing #

Thanks for helping make Twirp better! This is great!

First, if you have run into a bug, please file an issue. We try to get back to
issue reporters within a day or two. We might be able to help you right away.

If you'd rather not publicly discuss the issue, please email spencer@twitch.tv
and/or security@twitch.tv.

Issues are also a good place to present experience reports or requests for new
features.

If you'd like to make changes to Twirp, read on:

## Setup Requirements ##

You will need git, Go 1.9+, and Python 2.7 installed and on your system's path.
Install them however you feel.

## Developer Loop ##

Generally you want to make changes and run `make`, which will install all
dependencies we know about, build the core, and run all of the tests that we
have against all of the languages we support.

Most tests of the Go server are in `internal/twirptest/service_test.go`. Tests
of cross-language clients are in the [clientcompat](./clientcompat) directory.

## Contributing Code ##

Twirp uses github pull requests. Fork, hack away at your changes, run the test
suite with `make`, and submit a PR.

## Contributing Documentation ##

Twirp's docs are generated with [Docusaurus](https://docusaurus.io/). You can
safely edit anything inside the [docs](./docs) directory, adding new pages or
editing them. You can edit the sidebar by editing
[website/sidebars.json](./website/sidebars.json).

Then, to render your changes, run docusaurus's local server. To do this:

 1. [Install docusaurus on your machine](https://docusaurus.io/docs/en/installation.html).
 2. `cd website`
 3. `npm start`
 4. Navigate to http://localhost:3000/.

## Releasing Versions ##

Releasing versions is the responsibility of the core maintainers. Most people
don't need to know this stuff.

Twirp uses [Semantic versioning](http://semver.org/): `v<major>.<minor>.<patch>`.

 * Increment major if you're making a backwards-incompatible change.
 * Increment minor if you're adding a feature that's backwards-compatible.
 * Increment patch if you're making a bugfix.

To make a release, remember to update the version number in
[internal/gen/version.go](./internal/gen/version.go).

Twirp uses Github releases. To make a new release:
 1. Merge all changes that should be included in the release into the master
    branch.
 2. Update the version constant in `internal/gen/version.go`.
 3. Add a new commit to master with a message like "Version vX.X.X release".
 4. Tag the commit you just made: `git tag <version number>` and `git push
    origin --tags`
 5. Go to Github https://github.com/twitchtv/twirp/releases and
    "Draft a new release".
 6. Make sure to document changes, specially when upgrade instructions are
    needed.


## Code of Conduct
This project has adopted the [Amazon Open Source Code of Conduct](https://aws.github.io/code-of-conduct).
For more information see the [Code of Conduct FAQ](https://aws.github.io/code-of-conduct-faq) or contact
opensource-codeofconduct@amazon.com with any additional questions or comments.


## Licensing

See the [LICENSE](https://github.com/twitchtv/twirp/blob/master/LICENSE) file for our project's licensing. We will ask you to confirm the licensing of your contribution.

We may ask you to sign a [Contributor License Agreement (CLA)](http://en.wikipedia.org/wiki/Contributor_License_Agreement) for larger changes.
