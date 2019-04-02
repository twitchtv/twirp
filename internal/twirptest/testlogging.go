package twirptest

import (
	"log"
	"testing"
)

// testLogger creates a *log.Logger that writes log messages to the test's
// output. This makes log messages appear only if the test fails, and makes them
// align correctly for nested subtests.
func testLogger(t *testing.T) *log.Logger {
	return log.New(testWriter{t}, "", log.LstdFlags)
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(string(p))
	return len(p), nil
}
