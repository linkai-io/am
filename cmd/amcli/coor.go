package main

import (
	"context"
	"fmt"
	"os"

	"github.com/linkai-io/am/clients/coordinator"
)

func processCoor(args []string) {
	if len(args) < 2 {
		fmt.Printf("amcli - insufficent arguments for coor command must be one of: start\n\n")
		coorCmd.PrintDefaults()
		os.Exit(-1)
	}

	switch args[1] {
	case "-h", "--help", "help":
		fmt.Printf("amcli - group\n\n")
		coorCmd.PrintDefaults()
		os.Exit(-1)
	case "start":
		coorCmd.Parse(args[2:])
		startGroup()
	default:
		fmt.Printf("amcli - group - unknown cmd must be one of: json, pause, get, add, rem\n\n")
		orgCmd.PrintDefaults()
		os.Exit(-1)
	}
}

func startGroup() {
	fmt.Printf("%d %d\n", groupID, orgID)
	if groupID == -1 || orgID == -1 || userID == -1 {
		printExit("need gid/uid/oid for coordination")
	}

	coorClient := coordinator.New()
	if err := coorClient.Init([]byte(coorAddr)); err != nil {
		printErrExit("failed to connect to coordinator: %s\n", err)
	}
	ctx := context.Background()

	if err := coorClient.StartGroup(ctx, newUserContext(orgID, userID), groupID); err != nil {
		printErrExit("failed to start group: %s\n", err)
	}

	fmt.Printf("Successfully started scangroup: oid: %d gid: %d\n", orgID, groupID)
}
