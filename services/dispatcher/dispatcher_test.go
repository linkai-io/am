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
	webClient := amtest.MockWebDataService(1, 1)
	modClients := mockModules(t)
	wg := &sync.WaitGroup{}
	groups := mockGroups(1, groupCount, t)
	state := amtest.MockDispatcherState(wg, groups)
	portClient := amtest.MockPortScanService(1, 1, []int32{80, 443, 8080})

	dependentServices := &dispatcher.DependentServices{
		EventClient:    eventClient,
		SgClient:       sgClient,
		AddressClient:  addrClient,
		WebClient:      webClient,
		ModuleClients:  modClients,
		PortScanClient: portClient,
	}

	d := dispatcher.New(dependentServices, state)
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
	if portClient.AddGroupInvoked == false {
		t.Fatalf("error add group was not invoked")
	}

	if portClient.RemoveGroupInvoked == false {
		t.Fatalf("error remove group was not invoked")
	}

	if portClient.AnalyzeInvoked == false {
		t.Fatalf("error port scan was not invoked")
	}
}

func mockGroups(orgID, num int, t *testing.T) []*am.ScanGroup {
	groups := make([]*am.ScanGroup, num)
	for i := 0; i < num; i++ {
		groups[i] = amtest.CreateScanGroupOnly(orgID, i)
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
