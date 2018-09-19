package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/linkai-io/am/am"
)

var (
	orgCmd    *flag.FlagSet
	orgAddCmd *flag.FlagSet
	orgRemCmd *flag.FlagSet

	groupCmd *flag.FlagSet
	userCmd  *flag.FlagSet
	addrCmd  *flag.FlagSet
	coorCmd  *flag.FlagSet

	spawnCmd *flag.FlagSet
)

var (
	orgID  int
	userID int
	addr   string

	orgData *am.Organization
	orgAddr string

	groupData      *am.ScanGroup
	groupFile      string
	groupAddr      string
	groupInputFile string
	groupName      string
	groupID        int
	groupOID       int
	groupUID       int
	groupPause     bool
	groupResume    bool

	addrAddr  string
	addrInput string
	addrStart int64
	addrLimit int

	coorAddr string

	spawnType    string
	spawnNetwork string
	spawnCount   int
)

func init() {
	orgData = &am.Organization{}
	orgCmd = flag.NewFlagSet("org", flag.ExitOnError)
	orgCmd.StringVar(&addr, "addr", ":8383", "org server address")
	orgCmd.StringVar(&orgData.OrgName, "name", "test", "organization name")
	orgCmd.StringVar(&orgData.FirstName, "first", "first_name", "owner's first name")
	orgCmd.StringVar(&orgData.LastName, "last", "last_name", "owner's last name")
	orgCmd.StringVar(&orgData.OwnerEmail, "email", "", "owner's email")
	orgCmd.IntVar(&orgData.OrgID, "id", -1, "organization id")

	groupData = &am.ScanGroup{}
	groupCmd = flag.NewFlagSet("group", flag.ExitOnError)
	groupCmd.StringVar(&addr, "addr", ":8383", "org server address")
	groupCmd.StringVar(&groupFile, "file", "scangroup.json", "file with scan group details")
	groupCmd.StringVar(&groupName, "name", "", "scan group name")
	groupCmd.IntVar(&groupOID, "oid", -1, "org id to use for this scan group")
	groupCmd.IntVar(&groupUID, "uid", -1, "user id to use for this scan group")
	groupCmd.IntVar(&groupID, "gid", -1, "group id to use for this scan group")
	groupCmd.BoolVar(&groupPause, "pause", false, "include this argument to pause")
	groupCmd.BoolVar(&groupResume, "resume", false, "include this argument to resume")
	groupCmd.StringVar(&groupInputFile, "input", "s3:///tmp/ips.txt", "path to input file note s3:// becomes file:// if local")

	addrCmd = flag.NewFlagSet("addr", flag.ExitOnError)
	addrCmd.StringVar(&addr, "addr", ":8383", "address server address")
	addrCmd.StringVar(&groupName, "name", "", "scan group name")
	addrCmd.IntVar(&groupID, "gid", -1, "scan group name for these addresses")
	addrCmd.IntVar(&orgID, "oid", -1, "org id to use for this scan group's addresses")
	addrCmd.IntVar(&userID, "uid", -1, "user id to use for this scan group's addresses")
	addrCmd.StringVar(&addrInput, "input", "inputs.txt", "file containing host input")
	addrCmd.Int64Var(&addrStart, "start", 0, "address id to start get from")
	addrCmd.IntVar(&addrLimit, "limit", 10000, "limit number of records to return")

	userCmd = flag.NewFlagSet("user", flag.ExitOnError)

	coorCmd = flag.NewFlagSet("coor", flag.ExitOnError)
	coorCmd.StringVar(&addr, "addr", ":8383", "coordinator server address")
	coorCmd.IntVar(&groupID, "gid", -1, "scan group id to coordinate")
	coorCmd.IntVar(&orgID, "oid", -1, "org id to use for coordination")
	coorCmd.IntVar(&userID, "uid", -1, "user id to use for coordination")

	spawnCmd = flag.NewFlagSet("spawn", flag.ExitOnError)
	spawnCmd.StringVar(&spawnType, "type", "dispatcher", "system to spawn: (dispatcher, ns, web, port, brute, keyword)")
	spawnCmd.StringVar(&spawnNetwork, "net", "scripts_app_net", "network to add this container to: (scripts_app_net)")
	spawnCmd.IntVar(&spawnCount, "count", 1, "number of containers to spawn")
}

func main() {
	flag.Parse()
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println("./amcli org: (add, rem, get)")
		orgCmd.PrintDefaults()
		fmt.Println("./amcli group: (json, get, add, rem, pause)")
		groupCmd.PrintDefaults()
		fmt.Println("./amcli addr: (add, rem, get)")
		addrCmd.PrintDefaults()
		fmt.Println("./amcli coor: (start)")
		coorCmd.PrintDefaults()
		fmt.Println("insufficient arguments")
		fmt.Println("./amcli spawn: (type)")
		spawnCmd.PrintDefaults()
		os.Exit(-1)
	}

	switch os.Args[1] {
	case "org":
		processOrg(os.Args[1:])
	case "group":
		processGroup(os.Args[1:])
	case "addr":
		processAddr(os.Args[1:])
	case "coor":
		processCoor(os.Args[1:])
	case "spawn":
		processSpawn(os.Args[1:])
	default:
		printExit("unknown cmd, must be one of: org, group, user, addr, coor, spawn")
	}
}

func newUserContext(oid, uid int) am.UserContext {
	return &am.UserContextData{
		OrgID:  oid,
		UserID: uid,
	}
}

func systemContext() am.UserContext {
	return newUserContext(1, 1)
}

func printExit(message string) {
	fmt.Printf("%s\n", message)
	os.Exit(-1)
}

func printErrExit(message string, err error) {
	fmt.Printf(message, err.Error())
	os.Exit(-1)
}
