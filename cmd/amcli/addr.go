package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"google.golang.org/grpc"

	"github.com/linkai-io/am/clients/address"
	"github.com/linkai-io/am/pkg/inputlist"
)

func processAddr(args []string) {
	if len(args) < 2 {
		fmt.Printf("amcli - insufficent arguments for addr command must be one of: \n\n")
		addrCmd.PrintDefaults()
		os.Exit(-1)
	}

	switch args[1] {
	case "-h", "--help", "help":
		fmt.Printf("amcli - addr\n\n")
		addrCmd.PrintDefaults()
		os.Exit(-1)
	case "add":
		addrCmd.Parse(args[2:])
		addAddrs()
	case "rem":
		addrCmd.Parse(args[2:])
		removeAddrs()
	case "get":
		addrCmd.Parse(args[2:])
		getAddrs()
	default:
		fmt.Printf("amcli - addr - unknown cmd must be one of: add, rem, get\n\n")
		orgCmd.PrintDefaults()
		os.Exit(-1)
	}
}

func addAddrs() {
	if addr == "" {
		printExit("addr server address required")
	}

	if orgID == -1 || userID == -1 {
		printExit("error oid and uid are both required")
	}

	if groupName == "" && groupID == -1 {
		printExit("error group required, either name or gid")
	}

	addrFile, err := os.Open(addrInput)
	if err != nil {
		printErrExit("error opening address input file: %s\n", err)
	}

	addrs, errs := inputlist.ParseList(addrFile, 10000)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Printf("%s on line %d (%s)\n", err.Err, err.LineNumber, err.Line)
		}
		fmt.Printf("the following errors occurred, continue? ")
		var resp string
		fmt.Scanln(&resp)
		if resp == "y" || resp == "Y" {
			fmt.Printf("ignoring %d errors\n", len(errs))
		} else {
			printExit("quitting due to errors in input list")
		}
	}
	fmt.Printf("adding %d addreses\n", len(addrs))

	addrClient := address.New()
	ctx := context.Background()
	if err = addrClient.Init([]byte(addr)); err != nil {
		printErrExit("error connecting to server: %s\n", err)
	}
	sgAddrs := makeAddrs(addrs, orgID, userID, groupID)

	oid, count, err := addrClient.Update(ctx, newUserContext(orgID, userID), sgAddrs)
	if err != nil {
		printErrExit("error adding addresses %s", err)
	}

	fmt.Printf("Successfully added %d addresses for OrgID: %d\n", count, oid)
}

func makeAddrs(in map[string]struct{}, orgID, userID, groupID int) []*am.ScanGroupAddress {
	addrs := make([]*am.ScanGroupAddress, len(in))
	i := 0
	for addr := range in {
		addrs[i] = &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			DiscoveredBy:        "input_list",
			DiscoveryTime:       time.Now().UnixNano(),
			ConfidenceScore:     100.0,
			UserConfidenceScore: 0.0,
		}

		if inputlist.IsIP(addr) {
			addrs[i].IPAddress = addr
		} else {
			addrs[i].HostAddress = addr
		}
		i++
	}
	return addrs
}

func removeAddrs() {

}

func getAddrs() {
	grpc.EnableTracing = true

	var err error
	if addr == "" {
		printExit("addr server address required")
	}

	if orgID == -1 || userID == -1 {
		printExit("error oid and uid are both required")
	}

	if groupID == -1 {
		printExit("error gid required")
	}

	addrClient := address.New()
	ctx := context.Background()
	if err = addrClient.Init([]byte(addr)); err != nil {
		printErrExit("error connecting to server: %s\n", err)
	}

	filter := &am.ScanGroupAddressFilter{
		OrgID:   orgID,
		GroupID: groupID,
		Start:   addrStart,
		Limit:   addrLimit,
	}
	_, addresses, err := addrClient.Get(ctx, newUserContext(orgID, userID), filter)
	if err != nil {
		printErrExit("error getting addresses %#v\n", err)
	}
	data, err := json.MarshalIndent(addresses, "", "\t")
	fmt.Printf("%s\n", string(data))
}
