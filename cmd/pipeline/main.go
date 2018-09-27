package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/dispatcher"
	"github.com/linkai-io/am/services/module/ns"
)

func main() {
	orgID := 1
	userID := 1
	groupID := 1

	addrFile, err := os.Open("testdata/netflix.txt")
	if err != nil {
		log.Fatalf("error opening test file: %s\n", err)
	}

	addresses := amtest.AddrsFromInputFile(orgID, groupID, addrFile, nil)
	callCount := 0
	addrClient := &mock.AddressService{}
	addrClient.GetFn = func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (int, []*am.ScanGroupAddress, error) {
		if callCount == 0 {
			callCount++
			return orgID, addresses, nil
		}
		return orgID, nil, nil
	}
	// init NS module state system & NS module
	nsstate := amtest.MockNSState()
	dc := dnsclient.New([]string{"127.0.0.53:53"}, 3)
	nsModule := ns.New(dc, nsstate)
	if err := nsModule.Init(nil); err != nil {
		log.Fatalf("error initializing ns module: %s\n", err)
	}
	modules := make(map[am.ModuleType]am.ModuleService)
	modules[am.NSModule] = nsModule

	count := 0
	// init Dispatcher state system and DispatcherService
	disState := &mock.DispatcherState{}
	stateAddrs := make([]*am.ScanGroupAddress, 0)        // addresses stored in state
	stateHashes := make(map[string]*am.ScanGroupAddress) // hashes stored in state

	disState.PutAddressesFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses []*am.ScanGroupAddress) error {
		stateAddrs = append(stateAddrs, addresses...)
		for _, addr := range addresses {
			stateHashes[addr.AddressHash] = addr
			count++
		}
		log.Printf("TOTAL %d, len state: %d len hashes: %d\n", count, len(stateAddrs), len(stateHashes))
		return nil
	}

	disState.PutAddressMapFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses map[string]*am.ScanGroupAddress) error {
		for _, v := range addresses {
			stateAddrs = append(stateAddrs, v)
			stateHashes[v.AddressHash] = v
			count++
		}
		log.Printf("TOTAL %d, len state: %d len hashes: %d\n", count, len(stateAddrs), len(stateHashes))

		return nil
	}

	disState.FilterNewFn = func(ctx context.Context, orgID, scanGroupID int, addresses map[string]*am.ScanGroupAddress) (map[string]*am.ScanGroupAddress, error) {
		for k := range addresses {
			if _, exist := stateHashes[k]; exist {
				delete(addresses, k)
			}
		}
		return addresses, nil
	}

	disState.PopAddressesFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[string]*am.ScanGroupAddress, error) {
		newAddrs := make(map[string]*am.ScanGroupAddress)
		for _, addr := range stateAddrs {
			newAddrs[addr.AddressHash] = addr
			count--
		}

		// clear out addresses
		stateAddrs = make([]*am.ScanGroupAddress, 0)
		log.Printf("after pop, len: %d, %v\n", len(newAddrs), newAddrs)
		return newAddrs, nil
	}
	dispatcher := dispatcher.New(addrClient, modules, disState)
	dispatcher.Init(nil)

	ctx := context.Background()
	userContext := amtest.CreateUserContext(orgID, userID)

	// Run pipeline
	dispatcher.PushAddresses(ctx, userContext, groupID)

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			log.Printf("Active Addresses: %d, Groups: %d\n", dispatcher.GetActiveAddresses(), dispatcher.GetActiveGroups())
			if dispatcher.GetActiveGroups() == 0 && dispatcher.GetActiveAddresses() == 0 {
				for _, v := range stateHashes {
					log.Printf("%#v\n", v)
				}
				return
			}
		}
	}
}
