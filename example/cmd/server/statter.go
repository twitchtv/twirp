package main

import (
	"fmt"
	"io"
	"time"
)

type LoggingStatter struct {
	io.Writer
}

func (ls LoggingStatter) Inc(metric string, val int64, rate float32) error {
	_, err := fmt.Fprintf(ls, "incr %s: %d @ %f\n", metric, val, rate)
	return err
}

func (ls LoggingStatter) TimingDuration(metric string, val time.Duration, rate float32) error {
	_, err := fmt.Fprintf(ls, "time %s: %s @ %f\n", metric, val, rate)
	return err
}
