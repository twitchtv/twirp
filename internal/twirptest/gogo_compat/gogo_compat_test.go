package gogo_compat

import (
	"testing"

	"github.com/twitchtv/twirp/internal/descriptors"
)

func TestCompilation(t *testing.T) {
	// Test passes if this package compiles
}

func TestReflection(t *testing.T) {
	// Despite use of gogo, we should still be able to reflect on the service.
	var s Svc
	server := NewSvcServer(s, nil)
	fd, sd, err := descriptors.ServiceDescriptor(server)
	if err != nil {
		t.Fatalf("ServiceDescriptor err: %v", err)
	}
	if have, want := fd.GetPackage(), "twirp.internal.twirptest.gogo_compat"; have != want {
		t.Errorf("bad package, have=%q, want=%q", have, want)
	}
	if have, want := sd.GetName(), "Svc"; have != want {
		t.Errorf("bad service name, have=%q, want=%q", have, want)
	}
}
