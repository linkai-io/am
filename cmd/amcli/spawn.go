package main

import (
	"fmt"
	"os"
)

func processSpawn(args []string) {
	if len(args) < 2 {
		fmt.Printf("amcli - insufficent arguments for addr command must be one of: \n\n")
		spawnCmd.PrintDefaults()
		os.Exit(-1)
	}

	switch args[1] {
	case "-h", "--help", "help":
		fmt.Printf("amcli - spawn\n\n")
		spawnCmd.PrintDefaults()
		os.Exit(-1)
	case "add":
		spawnCmd.Parse(args[2:])
		addAddrs()
	case "rem":
		spawnCmd.Parse(args[2:])
		removeAddrs()
	default:
		fmt.Printf("amcli - spawn - unknown cmd must be one of: add, rem\n\n")
		spawnCmd.PrintDefaults()
		os.Exit(-1)
	}
}

func spawn() {
	name := "linkai_"
	switch spawnType {
	case "ns":
		name += "nsmoduleservice"
	case "dispatcher":
		name += "dispatcher"
	}

}

func kill() {

}

func getRandomPort() int {
	return 0
}
