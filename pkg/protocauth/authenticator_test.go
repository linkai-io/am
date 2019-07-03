package protocauth_test

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/linkai-io/am/pkg/protocauth"
)

func TestAuthenticate(t *testing.T) {
	a := protocauth.New([]byte("sometoken"))
	md := metadata.New(map[string]string{"authorization": "Bearer sometoken"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	newCtx, err := a.Authenticate(ctx)
	if err != nil {
		t.Fatalf("failed to authenticate: %v\n", err)
	}

	if newCtx == nil {
		t.Fatalf("returned context was nil")
	}
}
