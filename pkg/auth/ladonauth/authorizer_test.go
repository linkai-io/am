package ladonauth_test

import (
	"testing"

	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/amtest"

	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
)

func TestNewLadonAuthorizer(t *testing.T) {
	db := amtest.InitDB(t)
	sqlManager := ladonauth.NewPolicyManager(db, "pgx")
	if err := sqlManager.Init(); err != nil {
		t.Fatalf("error initializing sql manager: %s\n", err)
	}

	roleManager := ladonauth.NewRoleManager(db, "pgx")
	if err := roleManager.Init(); err != nil {
		t.Fatalf("error initialzing role manager: %s\n", err)
	}

	authorizer := ladonauth.NewLadonAuthorizer(sqlManager, roleManager)

	if err := authorizer.IsAllowed(am.EditorRole, am.RNScanGroupGroups, "create"); err == nil {
		t.Fatalf("editor role should not be allowed to create groups")
	}

	if err := authorizer.IsAllowed(am.AdminRole, am.RNScanGroupGroups, "create"); err != nil {
		t.Fatalf("error admin role should be allowed to create groups, got: %s\n", err)
	}
}
