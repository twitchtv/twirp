package logbodies

import (
	"context"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/internal/twirptest"
)

func serverAndClient(hooks *twirp.ServerHooks) (*httptest.Server, twirptest.Haberdasher) {
	return twirptest.ServerAndClient(twirptest.NoopHatmaker(), hooks)
}

func TestLogbodiesHooks(t *testing.T) {
	logger := log.New(os.Stderr, "test-logbodies", log.LstdFlags)
	hooks := LogBodies(logger)
	server, client := serverAndClient(hooks)
	defer server.Close()

	_, err := client.MakeHat(context.Background(), &twirptest.Size{Inches: 1})
	if err != nil {
		t.Fatalf("twirptest Client err=%q", err)
	}

}
