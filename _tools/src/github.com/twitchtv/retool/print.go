package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func log(msg string) {
	fmt.Fprintf(os.Stderr, "retool: %s\n", msg)
}

func fatal(msg string, err error) {
	if err == nil {
		fmt.Fprintf(os.Stderr, "retool: fatal err: %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "retool: fatal err: %s: %s\n", msg, err)
	}
	os.Exit(1)
}

func fatalExec(cmd string, err error) {
	if exErr, ok := errors.Cause(err).(*exec.ExitError); ok {
		fatal(fmt.Sprintf("execution error on %q: %s", cmd, exErr.Stderr), err)
	} else {
		fatal(fmt.Sprintf("execution error on %q", cmd), err)
	}
}
