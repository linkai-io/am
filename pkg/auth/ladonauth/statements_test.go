package ladonauth_test

import (
	"testing"

	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
)

func TestGetStatements(t *testing.T) {
	s := ladonauth.GetPolicyStatements("pgx")
	if s == nil {
		t.Fatalf("statements was nil")
	}

	s = ladonauth.GetPolicyStatements("blah")
	if s != nil {
		t.Fatalf("expected nil got: %#v\n", s)
	}
}

func TestGetRoleStatements(t *testing.T) {
	s := ladonauth.GetRoleStatements("pgx")
	if s == nil {
		t.Fatalf("statements was nil")
	}

	s = ladonauth.GetRoleStatements("blah")
	if s != nil {
		t.Fatalf("expected nil got: %#v\n", s)
	}
}
