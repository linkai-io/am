package dispatcher_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"

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
	callCount := 0
	addrClient := &mock.AddressService{}
	addrClient.GetFn = func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (int, []*am.ScanGroupAddress, error) {
		if callCount == 0 {
			callCount++
			return orgID, addresses, nil
		}
		return orgID, nil, nil
	}

	nsstate := amtest.MockNSState()
	nsModule := ns.New(nsstate)
	if err := nsModule.Init(nil); err != nil {
		t.Fatalf("error initializing ns module: %s\n", err)
	}
	modules := make(map[am.ModuleType]am.ModuleService)
	modules[am.NSModule] = nsModule

	disState := &mock.DispatcherState{}
	stateAddrs := make([]*am.ScanGroupAddress, 0)
	disState.PutAddressesFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses []*am.ScanGroupAddress) error {
		stateAddrs = append(stateAddrs, addresses...)
		return nil
	}

	disState.GetAddressesFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[int64]*am.ScanGroupAddress, error) {
		newAddrs := make(map[int64]*am.ScanGroupAddress)
		for _, addr := range stateAddrs {
			newAddrs[addr.AddressID] = addr
		}
		// clear out addresses
		stateAddrs = make([]*am.ScanGroupAddress, 0)
		return newAddrs, nil
	}
	dispatcher := dispatcher.New(addrClient, modules, disState)
	dispatcher.Init(nil)

	ctx := context.Background()
	userContext := amtest.CreateUserContext(orgID, userID)
	dispatcher.PushAddresses(ctx, userContext, groupID)
	time.Sleep(5 * time.Second)
}
