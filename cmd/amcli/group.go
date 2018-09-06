package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/scangroup"
)

func processGroup(args []string) {
	if len(args) < 2 {
		fmt.Printf("amcli - insufficent arguments for group command must be one of: json, get, add, rem\n\n")
		groupCmd.PrintDefaults()
		os.Exit(-1)
	}

	switch args[1] {
	case "-h", "--help", "help":
		fmt.Printf("amcli - group\n\n")
		groupCmd.PrintDefaults()
		os.Exit(-1)
	case "json":
		groupCmd.Parse(args[2:])
		generateJSON()
	case "add":
		groupCmd.Parse(args[2:])
		addGroup()
	case "rem":
		groupCmd.Parse(args[2:])
		removeGroup()
	case "get":
		groupCmd.Parse(args[2:])
		getGroup()
	case "pause":
		groupCmd.Parse(args[2:])
		pauseGroup()
	default:
		fmt.Printf("amcli - group - unknown cmd must be one of: json, pause, get, add, rem\n\n")
		orgCmd.PrintDefaults()
		os.Exit(-1)
	}
}

func generateJSON() {
	groupData.CreatedBy = 1
	groupData.CreationTime = time.Now().UnixNano()
	groupData.Deleted = false
	groupData.GroupID = 1
	groupData.GroupName = "groupName"
	groupData.ModifiedBy = 1
	groupData.ModifiedTime = time.Now().UnixNano()
	groupData.ModuleConfigurations = &am.ModuleConfiguration{
		NSModule: &am.NSModuleConfig{
			RequestsPerSecond: 50,
		},
		BruteModule: &am.BruteModuleConfig{
			RequestsPerSecond: 50,
			CustomSubNames:    []string{"custom"},
			MaxDepth:          2,
		},
		PortModule: &am.PortModuleConfig{
			RequestsPerSecond: 50,
			CustomPorts:       []int32{8080, 8443},
		},
		WebModule: &am.WebModuleConfig{
			RequestsPerSecond:     50,
			TakeScreenShots:       true,
			MaxLinks:              2,
			ExtractJS:             true,
			FingerprintFrameworks: true,
		},
		KeywordModule: &am.KeywordModuleConfig{
			Keywords: []string{"keywords"},
		},
	}
	groupData.OrgID = 1
	groupData.OriginalInputS3URL = "s3://bucket/org/file"
	groupData.Paused = false

	data, err := json.MarshalIndent(groupData, "", "\t")
	if err != nil {
		printErrExit("error marshaling scan group data: %s\n", err)
	}
	fmt.Printf("%s\n", string(data))
}

func addGroup() {
	if groupAddr == "" {
		printExit("group server address is required")
	}

	if groupOID == -1 || groupUID == -1 {
		printExit("error oid and uid are required")
	}

	if groupName == "" {
		printExit("error name is required")
	}

	r, err := os.Open(groupFile)
	if err != nil {
		printErrExit("error opening file for reading: %s\n", err)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		printErrExit("error reading file: %s\n", err)
	}

	if err := json.Unmarshal(data, &groupData); err != nil {
		printErrExit("error reading data: %s\n", err)
	}

	groupData.OrgID = groupOID
	groupData.CreatedBy = groupUID
	groupData.ModifiedBy = groupUID
	groupData.OriginalInputS3URL = groupInputFile
	groupData.GroupName = groupName
	groupData.Paused = groupPause

	groupClient := scangroup.New()
	ctx := context.Background()
	if err = groupClient.Init([]byte(groupAddr)); err != nil {
		printErrExit("error connecting to server: %s\n", err)
	}

	oid, gid, err := groupClient.Create(ctx, newUserContext(groupOID, groupUID), groupData)
	if err != nil {
		printErrExit("failed to create group: %s\n", err)
	}
	fmt.Printf("Successfully created new scangroup: oid: %d gid: %d\n", oid, gid)
}

func removeGroup() {
	var oid int
	var gid int
	var err error

	if groupAddr == "" {
		printExit("group server address is required")
	}

	if groupOID == -1 || groupUID == -1 {
		printExit("error oid and uid are required")
	}

	if groupID == -1 && groupName == "" {
		printExit("error gid or name must be set")
	}

	groupClient := scangroup.New()
	ctx := context.Background()
	if err = groupClient.Init([]byte(groupAddr)); err != nil {
		printErrExit("error connecting to server: %s\n", err)
	}

	if groupID != -1 {
		oid, gid, err = groupClient.Delete(ctx, newUserContext(groupOID, groupUID), groupID)
	} else {
		oid, groupData, err = groupClient.GetByName(ctx, newUserContext(groupOID, groupUID), groupName)
		if err != nil {
			printErrExit("error getting group by name: %s\n", err)
		}
		oid, gid, err = groupClient.Delete(ctx, newUserContext(groupOID, groupUID), groupData.GroupID)
	}

	if err != nil {
		printErrExit("failed to remove group: %s\n", err)
	}

	fmt.Printf("Successfully removed group for OrgID: %d groupID: %d\n", oid, gid)
}

func getGroup() {
	var err error

	if groupAddr == "" {
		printExit("group server address is required")
	}

	if groupOID == -1 || groupUID == -1 {
		printExit("error oid and uid are required")
	}

	if groupID == -1 && groupName == "" {
		printExit("error gid or name must be set")
	}

	groupClient := scangroup.New()
	ctx := context.Background()
	if err = groupClient.Init([]byte(groupAddr)); err != nil {
		printErrExit("error connecting to server: %s\n", err)
	}

	if groupID != -1 {
		_, groupData, err = groupClient.Get(ctx, newUserContext(groupOID, groupUID), groupID)
	} else {
		_, groupData, err = groupClient.GetByName(ctx, newUserContext(groupOID, groupUID), groupName)
		if err != nil {
			printErrExit("error getting group by name: %s\n", err)
		}
	}
	data, err := json.MarshalIndent(groupData, "", "\t")
	if err != nil {
		printErrExit("unable to unmarshal scan group: %s", err)
	}
	fmt.Printf("%s\n", string(data))
}

func pauseGroup() {
	var oid int
	var gid int
	var err error

	if groupAddr == "" {
		printExit("group server address is required")
	}

	if groupOID == -1 || groupUID == -1 {
		printExit("error oid and uid are required")
	}

	if groupID == -1 && groupName == "" {
		printExit("error gid or name must be set")
	}
	if groupPause == false && groupResume == false {
		printExit("error either pause or resume must be set")
	}

	groupClient := scangroup.New()
	ctx := context.Background()
	if err = groupClient.Init([]byte(groupAddr)); err != nil {
		printErrExit("error connecting to server: %s\n", err)
	}

	if groupID != -1 {
		_, groupData, err = groupClient.Get(ctx, newUserContext(groupOID, groupUID), groupID)
	} else {
		_, groupData, err = groupClient.GetByName(ctx, newUserContext(groupOID, groupUID), groupName)
		if err != nil {
			printErrExit("error getting group by name: %s\n", err)
		}
		groupID = groupData.GroupID
	}

	if groupPause {
		fmt.Printf("Pausing groupID: %d\n", groupID)
		oid, gid, err = groupClient.Pause(ctx, newUserContext(groupOID, groupUID), groupID)
	}

	if groupResume {
		fmt.Printf("Resume groupID: %d\n", groupID)
		oid, gid, err = groupClient.Resume(ctx, newUserContext(groupOID, groupUID), groupID)
	}

	if err != nil {
		printErrExit("error updating client: %s\n", err)
	}
	fmt.Printf("Successfully updated Group: %d for OrgID: %d\n", gid, oid)
}
