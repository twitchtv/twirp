package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Masterminds/semver"
)

var version = semver.MustParse("v1.3.5")

func main() {
	flag.Parse()
	if err := ensureTooldir(); err != nil {
		fatal("failed to locate or create tool directory", err)
	}
	cmd, tool := parseArgs()

	if cmd == "version" {
		fmt.Fprintf(os.Stdout, "retool %s", version)
		os.Exit(0)
	}

	if !specExists() {
		if cmd == "add" {
			err := writeBlankSpec()
			if err != nil {
				fatal("failed to write blank spec", err)
			}
		} else {
			fatal("tools.json does not yet exist. You need to add a tool first with 'retool add'", nil)
		}
	}

	s, err := read()
	if err != nil {
		fatal("failed to load tools.json", err)
	}

	switch cmd {
	case "add":
		s.add(tool)
	case "upgrade":
		s.upgrade(tool)
	case "remove":
		s.remove(tool)
	case "build":
		s.build()
	case "sync":
		s.sync()
	case "do":
		s.sync()
		do()
	case "clean":
		log("the clean subcommand is deprecated and has no effect")
	default:
		fatal(fmt.Sprintf("unknown cmd %q", cmd), nil)
	}
}
