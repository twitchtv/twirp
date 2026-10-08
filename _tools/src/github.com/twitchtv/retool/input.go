package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	verboseFlag = flag.Bool("verbose", false, "Enable more detailed output that may be helpful for troubleshooting.")
	forkFlag    = flag.String("f", "", "Use a fork of the repository rather than the default upstream")

	// TODO: Refactor so that this global state is not necessary.
	positionalArgs []string
)

func verbosef(format string, a ...interface{}) {
	if *verboseFlag {
		_, _ = fmt.Fprintf(os.Stderr, format, a...)
	}
}

func parseArgs() (command string, t *tool) {
	if !flag.Parsed() {
		panic("parseArgs expects that flags have already been parsed")
	}

	args := flag.Args()

	if len(args) < 1 {
		printUsageAndExit("", 1)
	}

	command = args[0]
	args = args[1:]
	t = new(tool)
	positionalArgs = args

	switch command {
	case "version":
		assertArgLength(args, command, 0)
		return "version", t

	case "sync":
		assertArgLength(args, command, 0)
		return "sync", t

	case "add":
		assertArgLength(args, command, 2)
		t.Repository = args[0]
		t.ref = args[1]
		t.Fork = *forkFlag
		return "add", t

	case "upgrade":
		assertArgLength(args, command, 2)
		t.Repository = args[0]
		t.ref = args[1]
		t.Fork = *forkFlag
		return "upgrade", t

	case "remove":
		assertArgLength(args, command, 1)
		t.Repository = args[0]
		return "remove", t

	case "do":
		// A variable number of arguments are permissible for the 'do' subcommand; they are passed via t.PositionalArgs.
		return "do", t

	case "clean":
		assertArgLength(args, command, 0)
		return "clean", t

	case "build":
		assertArgLength(args, command, 0)
		return "build", t

	case "help":
		assertArgLength(args, command, 1)
		printUsageAndExit(args[0], 0)

	default:
		printUsageAndExit("", 1)
	}
	return "", t
}

func assertArgLength(args []string, command string, arglength int) {
	if len(args) != arglength {
		printUsageAndExit(command, 1)
	}
}

func printUsageAndExit(command string, exitCode int) {
	switch command {
	case "add":
		fmt.Println(addUsage)
	case "remove":
		fmt.Println(removeUsage)
	case "upgrade":
		fmt.Println(upgradeUsage)
	case "sync":
		fmt.Println(syncUsage)
	case "do":
		fmt.Println(doUsage)
	case "clean":
		fmt.Println(cleanUsage)
	case "build":
		fmt.Println(buildUsage)
	default:
		fmt.Println(usage)
	}
	os.Exit(exitCode)
}

const usage = `usage: retool (add | remove | upgrade | sync | do | build | help)

use retool with a subcommand:

add will add a tool
remove will remove a tool
upgrade will upgrade a tool
sync will synchronize your _tools with tools.json, downloading if necessary
build will compile all the tools in _tools
do will run stuff using your installed tools

help [command] will describe a command in more detail
version will print the installed version of retool
`

const addUsage = `usage: retool add [repository] [commit]

eg: retool add github.com/tools/godep 3020345802e4bff23902cfc1d19e90a79fae714e

Add will mark a repository as a tool you want to use. It will rewrite
tools.json to record this fact. It will then fetch the repository,
reset it to the desired commit, and install it to _tools/bin.

You can also use a symbolic reference, like 'master' or
'origin/master' or 'origin/v1.0'. Retool will end up parsing this and
storing the underlying SHA.
`

const upgradeUsage = `usage: retool upgrade [repository] [commit]

eg: retool upgrade github.com/tools/godep 3020345802e4bff23902cfc1d19e90a79fae714e

Upgrade set the commit SHA of a tool you want to use. It will
rewrite tools.json to record this fact. It will then fetch the
repository, reset it to the desired commit, and install it to
_tools/bin.

You can also use a symbolic reference, like 'master' or
'origin/master' or 'origin/v1.0'. Retool will end up parsing this and
storing the underlying SHA.
`

const removeUsage = `usage: retool remove [repository]

eg: retool remove github.com/tools/godep

Remove will remove a tool from your tools.json. It won't delete the
underlying repo from _tools, because it might be a dependency of some
other tool. If you want to clean things up, retool sync will clear out
unused dependencies.
`

const syncUsage = `usage: retool sync

Sync will synchronize your _tools directory to match tools.json. It will do this by making network
calls to download tools and set them to the right versions.

If you want to just install whatever is in _tools/src without using network, see 'retool build'.
`

const doUsage = `usage: retool do [command and args]

retool do will make sure your _tools directory is synced, and then
execute a command with the tools installed in _tools.

This is just
  retool sync && PATH=$PWD/_tools/bin:$PATH [command and args]
That works too.
`

const cleanUsage = `usage: retool clean

retool clean has no effect, but still exists for compatibility.
`

const buildUsage = `usage: retool build

retool build will compile all the tools listed in tools.json, obeying whatever is currently
downloaded into _tools. It will not do additional network calls. This is typically useful for
compiling vendored tools so you can use them inside isolated environments.

retool sync is the more full-featured version of retool build. It will actually do git fetches to
make sure the right stuff gets installed, but this requires network access.
`
