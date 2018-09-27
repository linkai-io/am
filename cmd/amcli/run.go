package main

func runCmd() {
	orgData.OrgName = "test"
	orgData.FirstName = "first_name"
	orgData.LastName = "last_name"
	orgData.OwnerEmail = "test@test.com"
	groupFile = "scangroup.json"
	groupName = "test_group"
	addrInput = "hosts.txt"
	addr = ":8383"

	orgID, userID = addOrg()
	groupID = addGroup()
	addAddrs()
	startGroup()
}
