package scangroup_test

import (
	"os"
	"testing"

	"gopkg.linkai.io/v1/repos/am/amtest"
	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
	"gopkg.linkai.io/v1/repos/am/services/scangroup"
)

func TestNew(t *testing.T) {
	db := amtest.InitDB(t)
	policyManager := ladonauth.NewPolicyManager(db, "pgx")
	roleManager := ladonauth.NewRoleManager(db, "pgx")
	auth := ladonauth.NewLadonAuthorizer(policyManager, roleManager)
	service := scangroup.New(auth)
	service.Init(os.Getenv("TEST"))

}
