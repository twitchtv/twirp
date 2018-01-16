package twirp

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pkg/errors"
)

func TestWithMetaRaces(t *testing.T) {
	err := NewError(Internal, "msg")
	err = err.WithMeta("k1", "v1")

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			_ = err.WithMeta(fmt.Sprintf("k-%d", i), "v")
			wg.Done()
		}(i)
	}

	wg.Wait()

	if len(err.MetaMap()) != 1 {
		t.Errorf("err was mutated")
	}
}

func TestErrorCause(t *testing.T) {
	rootCause := fmt.Errorf("this is only a test")
	twerr := InternalErrorWith(rootCause)
	cause := errors.Cause(twerr)
	if cause != rootCause {
		t.Errorf("got wrong cause for err. have=%q, want=%q", cause, rootCause)
	}
}
