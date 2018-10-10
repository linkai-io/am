package main

import (
	"os"

	"github.com/linkai-io/am/pkg/generators"
)

func runCmd() {
	orgData.OrgName = "test" + generators.InsecureAlphabetString(8)
	orgData.FirstName = "first_name"
	orgData.LastName = "last_name"
	orgData.OwnerEmail = "test@" + orgData.OrgName + ".com"
	groupFile = "scangroup.json"
	groupName = "test_group"
	addrInput = os.Args[2]
	addr = ":8383"

	orgID, userID = addOrg()
	groupID = addGroup()
	addAddrs()
	startGroup()
}
