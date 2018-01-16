package gen

import (
	"log"
	"os"
	"strings"
)

func Fail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("error:", s)
	os.Exit(1)
}

func Error(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ":" + err.Error()
	log.Print("error:", s)
	os.Exit(1)
}
