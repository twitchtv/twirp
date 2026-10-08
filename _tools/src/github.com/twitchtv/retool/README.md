# Retool: Vendor thy tools! #

[![Build Status](https://travis-ci.org/twitchtv/retool.svg?branch=master)](https://travis-ci.org/twitchtv/retool)
[![Windows Build status](https://ci.appveyor.com/api/projects/status/06x2vd5nh683iscu/branch/master?svg=true&passingText=Windows%20build%20passing&pendingText=Windows%20build%20pending&failingText=Windows%20build%20failing)](https://ci.appveyor.com/project/spenczar/retool/branch/master)


## what is this ##

retool helps manage the versions of _tools_ that you use with your
repository. These are executables that are a crucial part of your
development environment, but aren't imported by any of your code, so
they don't get scooped up by glide or godep (or any other vendoring
tool).

Some examples of tools:

 - [github.com/tools/godep](https://github.com/tools/godep) is a tool to
   vendor Go packages.
 - [github.com/golang/protobuf/protoc-gen-go](https://github.com/golang/protobuf/protoc-gen-go)
   is a tool to compile Go code from protobuf definitions.
 - [github.com/maxbrunsfeld/counterfeiter](https://github.com/maxbrunsfeld/counterfeiter)
   is a tool to generate mocks of interfaces.

You want this if you use code generation: if everybody has a different
version of the code generator, then you'll get meaningless churn across
runs of the generator unless everyone is pinned to the right version.

You might also want this if you use linters or tools like
[github.com/kisielk/errcheck](https://github.com/kisielk/errcheck) in an
automated fashion and you want to make sure that everyone has the same
version so you can pass flags to the linter with confidence.

retool pins on a per-project basis. It works by making a complete GOPATH
within your project. You can choose to commit the source files for those
tools, if you like.

## usage ##

The expected workflow is something like this:

Install retool:
```sh
go get github.com/twitchtv/retool
```

Add a tool dependency:
```sh
retool add github.com/jteeuwen/go-bindata/go-bindata origin/master
```

Use it to generate code:
```sh
retool do go-bindata -pkg testdata -o ./testdata/testdata.go ./testdata/data.json
```

---

There are a few other commands that you'll use much less often:

Upgrade a tool to its latest version:
```sh
retool upgrade github.com/spf13/hugo origin/master
# or to a particular tag
retool upgrade github.com/spf13/hugo v0.17
```

Stop using that tool you dont like anymore:
```sh
retool remove github.com/tools/godep
```

Compile all the tools that other people have vendored in a project:
```sh
# compiles everything without using the network - useful for isolated build environments
retool build
```

Double-check that you're in sync by comparing everything with upstream versions:
```sh
# makes sure your tools match tools.json by comparing against their remotes
retool sync
```

## why would i need to manage these things  ##

**TL;DR:** if you work with anyone else on your project, and they have
different versions of their tools, everything turns to shit.

One of the best parts about Go is that it is very, very simple. This
makes it straightforward to write code generation utilities. You don't
need to generate code for every project, but in large ones, code
generation can help you be much more productive.

Like, if you're writing tests that use an interface, you can use code
generation to quickly whip up structs which mock the interface so you
can force them to return errors. This way, you can test edge cases for
your interaction points with
interfaces. [github.com/maxbrunsfeld/counterfeiter](https://github.com/maxbrunsfeld/counterfeiter)
does this pretty well!

If you want to use the generated code, you should check in the
generated `.go` code to git, not just the sources, so that build boxes
and the like don't need all these code generation tools, and so that
`go get` just works cleanly.

This poses a problem, though, as soon as you start working with other
people on your project: if you have different versions of your code
generation tools, which generate slightly different output, you'll get
lots of meaningless churn in your commits. This sucks! There has to be
a better way!

## the retool way ##

retool records the versions of tools you want in a file, `tools.json`.
The file looks like this:

```json
{
  "Tools": [
    {
      "Repository": "github.com/golang/protobuf/protoc-gen-go",
      "Commit": "2fea9e168bab814ca0c6e292a6be164f624fc6ca"
    }
  ]
}
```

Tools are identified by repo and commit. Each tool in `tools.json` will
be installed to `_tools`, which is a private GOPATH just dedicated to
keeping track of these tools.

In practice, you don't need to know much about `tools.json`. You check
it in to git so that everybody stays in sync, but you manage it with
`retool add|upgrade|remove`.

When it's time to generate code, **instead of `go generate ./...`**, you
use `retool do go generate ./...` to use your sweet, vendored tools.
This really just calls `PATH=$PWD/_tools/bin:PATH go generate ./...`; if
you want to do anything fancy, you can feel free to use that path too.

## contributing to retool ##

Any pull requests are extremely welcome! If you run into problems or
have questions, please raise a github issue!

Retool's tests are mostly integration tests. They require a working Go
compiler, a working version of git, and network access.
