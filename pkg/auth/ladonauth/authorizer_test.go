package ladonauth_test

import (
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/pkg/auth/ladonauth"
)

func TestNewLadonAuthorizer(t *testing.T) {
	db := amtest.InitDB(env, t)
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
