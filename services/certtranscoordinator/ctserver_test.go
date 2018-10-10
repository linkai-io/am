package certtranscoordinator_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/services/certtranscoordinator"
)

func TestCTServers(t *testing.T) {
	servers := certtranscoordinator.NewCTServers()
	ctx := context.Background()
	list := servers.UpdateServers(ctx)
	if servers.Len() != 0 {
		t.Fatalf("server len should be zero")
	}
	if len(list) == 0 {
		t.Fatalf("server len was zero")
	}
	for _, server := range list {
		servers.Return(server)
	}

	if len(list) != servers.Len() {
		t.Fatalf("server lislt size did not match servers in ctserver got %d expected %d\n", servers.Len(), len(list))
	}

}
