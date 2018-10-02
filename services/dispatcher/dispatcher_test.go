package dispatcher_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/dnsclient"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/services/module/ns"

	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/services/dispatcher"
)

func TestDispatcherFlow(t *testing.T) {
	orgID := 1
	userID := 1
	groupID := 1

	addrFile, err := os.Open("testdata/netflix.txt")
	if err != nil {
		t.Fatalf("error opening test file: %s\n", err)
	}

	addresses := amtest.AddrsFromInputFile(orgID, groupID, addrFile, t)
	addrFile.Close()
	
	callCount := 0
	addrClient := &mock.AddressService{}
	addrClient.GetFn = func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (int, []*am.ScanGroupAddress, error) {
		if callCount == 0 {
			callCount++
			return orgID, addresses, nil
		}
		return orgID, nil, nil
	}

	addrClient.UpdateFn = func(ctx context.Context, userContext am.UserContext, addresses map[string]*am.ScanGroupAddress) (int, int, error) {
		return orgID, len(addresses), nil
	}
	// init NS module state system & NS module
	nsstate := amtest.MockNSState()
	dc := dnsclient.New([]string{"127.0.0.53:53"}, 3)
	nsModule := ns.New(dc, nsstate)
	if err := nsModule.Init(nil); err != nil {
		t.Fatalf("error initializing ns module: %s\n", err)
	}
	modules := make(map[am.ModuleType]am.ModuleService)
	modules[am.NSModule] = nsModule

	count := 0
	wg := &sync.WaitGroup{}
	wg.Add(1)

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
		t.Logf("TOTAL %d, len state: %d len hashes: %d\n", count, len(stateAddrs), len(stateHashes))
		return nil
	}

	disState.PutAddressMapFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses map[string]*am.ScanGroupAddress) error {
		for _, v := range addresses {
			stateAddrs = append(stateAddrs, v)
			stateHashes[v.AddressHash] = v
			count++
		}
		t.Logf("TOTAL %d, len state: %d len hashes: %d\n", count, len(stateAddrs), len(stateHashes))

		return nil
	}

	disState.FilterNewFn = func(ctx context.Context, orgID, scanGroupID int, addresses map[string]*am.ScanGroupAddress) (map[string]*am.ScanGroupAddress, error) {
		for k, _ := range addresses {
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
		t.Logf("after pop, len: %d, %v\n", len(newAddrs), newAddrs)
		return newAddrs, nil
	}

	disState.StopFn = func(ctx context.Context, userContext am.UserContext, scanGroupId int) error {
		wg.Done()
		return nil
	}

	dispatcher := dispatcher.New(addrClient, modules, disState)
	dispatcher.Init(nil)

	ctx := context.Background()
	userContext := amtest.CreateUserContext(orgID, userID)

	// Run pipeline
	dispatcher.PushAddresses(ctx, userContext, groupID)

	wg.Wait()
}

func TestTime(t *testing.T) {
	now := time.Now()
	var nowN int64
	// TODO: do smart calculation on size of scan group addresses
	then := now.Add(-1 * time.Minute).UnixNano()
	nowN = 1538094686180867747 //now.UnixNano() //1538093557888605642
	fmt.Printf("%d\n", nowN)
	fmt.Printf("%d\n", then)
	if nowN < then {
		fmt.Printf("scanning!")
	}
}
