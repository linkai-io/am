package dispatcher_test

import (
	"context"
	"sync"
	"testing"

	"github.com/linkai-io/am/mock"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/services/dispatcher"
)

func TestPushAddresses(t *testing.T) {
	groupCount := 10
	ctx := context.Background()
	sgClient := amtest.MockScanGroupService(1, 1)
	eventClient := amtest.MockEventService()
	addrs := amtest.GenerateAddrs(1, 1, 1)
	addrClient := amtest.MockAddressService(1, addrs)
	modClients := mockModules(t)
	wg := &sync.WaitGroup{}
	groups := mockGroups(1, groupCount, t)
	state := amtest.MockDispatcherState(wg, groups)
	d := dispatcher.New(sgClient, eventClient, addrClient, modClients, state)
	if err := d.Init(nil); err != nil {
		t.Fatalf("error initializing dispatcher")
	}
	for i := 0; i < groupCount; i++ {
		wg.Add(1)
		go func(i int) {
			if err := d.PushAddresses(ctx, amtest.CreateUserContext(1, 1), i); err != nil {
				t.Fatalf("error pushing addresses: %v\n", err)
			}
		}(i)
		t.Logf("pushed addr")
	}
	wg.Wait()
	defer d.Stop(ctx)
}

func mockGroups(orgID, num int, t *testing.T) []*am.ScanGroup {
	groups := make([]*am.ScanGroup, num)
	for i := 0; i < num; i++ {
		groups[i] = amtest.BuildScanGroup(orgID, i)
	}
	return groups
}

func mockModules(t *testing.T) map[am.ModuleType]am.ModuleService {
	modClients := make(map[am.ModuleType]am.ModuleService)
	for i := 0; i < 7; i++ {
		mod := &mock.ModuleService{}
		mod.AnalyzeFn = func(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
			t.Logf("module %s analyze called", am.KeyFromModuleType(am.ModuleType(i)))
			return address, nil, nil
		}
		modClients[am.ModuleType(i)] = mod
	}
	return modClients
}
