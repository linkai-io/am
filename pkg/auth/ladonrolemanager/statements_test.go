package ladonrolemanager_test

import (
	"testing"

	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonrolemanager"
)

func TestGetStatements(t *testing.T) {
	s := ladonrolemanager.GetStatements("pgx")
	if s == nil {
		t.Fatalf("statements was nil")
	}

	s = ladonrolemanager.GetStatements("blah")
	if s != nil {
		t.Fatalf("expected nil got: %#v\n", s)
	}
}
