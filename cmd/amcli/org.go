package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/organization"
)

func processOrg(args []string) {
	//orgCmd.Parse(args)
	if len(args) < 2 {
		fmt.Printf("amcli - insufficent arguments for org command\n\n")
		orgCmd.PrintDefaults()
		os.Exit(-1)
	}

	switch args[1] {
	case "-h", "--help", "help":
		fmt.Printf("amcli - org\n\n")
		orgCmd.PrintDefaults()
		os.Exit(-1)
	case "add":
		orgCmd.Parse(args[2:])
		addOrg()
	case "rem":
		orgCmd.Parse(args[2:])
		removeOrg()
	case "get":
		orgCmd.Parse(args[2:])
		getOrg()
	default:
		fmt.Printf("amcli - org - unknown cmd must be one of: get, add, rem\n\n")
		orgCmd.PrintDefaults()
		os.Exit(-1)
	}
}

func getOrg() {
	var err error
	var org *am.Organization

	if orgAddr == "" {
		printExit("org server address must not be empty")
	}

	if orgData.OrgName == "" || orgData.OrgID == -1 {
		printExit("remove requires org name or org id")
	}

	orgClient := organization.New()
	ctx := context.Background()
	if err = orgClient.Init([]byte(orgAddr)); err != nil {
		printExit(fmt.Sprintf("error connecting to server: %s\n", err))
	}

	if orgData.OrgID != -1 {
		_, org, err = orgClient.GetByID(ctx, systemContext(), orgData.OrgID)
		if err != nil {
			printExit(fmt.Sprintf("error deleting org: %s\n", err))
		}
	} else {
		_, org, err = orgClient.Get(ctx, systemContext(), orgData.OrgName)
		if err != nil {
			printExit(fmt.Sprintf("error getting org by name %s: %s\n", orgData.OrgName, err))
		}
	}
	fmt.Printf("%#v\n", org)

}
func addOrg() {
	if orgAddr == "" {
		printExit("org server address must not be empty")
	}

	if orgData.OrgName == "" {
		printExit("org name must not be empty")
	}

	if orgData.LastName == "" {
		printExit("last must not be empty")
	}

	if orgData.FirstName == "" {
		printExit("first must not be empty")
	}

	if orgData.OwnerEmail == "" {
		printExit("email must not be empty")
	}

	orgData.Address1 = "address1"
	orgData.Address2 = "address2"
	orgData.City = "city"
	orgData.Country = "country"
	orgData.CreationTime = time.Now().UnixNano()
	orgData.IdentityPoolID = "identity.pool"
	orgData.Phone = "phone"
	orgData.PostalCode = "postal"
	orgData.StatePrefecture = "state"
	orgData.StatusID = am.OrgStatusActive
	orgData.Street = "street"
	orgData.SubscriptionID = am.SubscriptionMonthly
	orgData.UserPoolID = "user.pool"

	orgClient := organization.New()
	ctx := context.Background()
	if err := orgClient.Init([]byte(orgAddr)); err != nil {
		printExit(fmt.Sprintf("error connecting to server: %s\n", err))
	}
	oid, uid, _, _, err := orgClient.Create(ctx, systemContext(), orgData)
	if err != nil {
		printExit(fmt.Sprintf("error creating org: %s\n", err))
	}
	fmt.Printf("Successfully created organization; OrgID: %d UserID: %d\n", oid, uid)
}

func removeOrg() {
	var err error
	var oid int
	if orgAddr == "" {
		printExit("org server address must not be empty")
	}

	if orgData.OrgName == "" || orgData.OrgID == -1 {
		printExit("remove requires org name or org id")
	}

	orgClient := organization.New()
	ctx := context.Background()
	if err = orgClient.Init([]byte(orgAddr)); err != nil {
		printExit(fmt.Sprintf("error connecting to server: %s\n", err))
	}

	if orgData.OrgID != -1 {
		oid, err = orgClient.Delete(ctx, systemContext(), orgData.OrgID)
		if err != nil {
			printExit(fmt.Sprintf("error deleting org: %s\n", err))
		}
	} else {
		orgID, _, err := orgClient.Get(ctx, systemContext(), orgData.OrgName)
		if err != nil {
			printExit(fmt.Sprintf("error getting org by name %s: %s\n", orgData.OrgName, err))
		}
		oid, err = orgClient.Delete(ctx, systemContext(), orgID)
		if err != nil {
			printExit(fmt.Sprintf("error deleting org: %s\n", err))
		}
	}
	fmt.Printf("Successfully removed OrgID: %d\n", oid)
}