package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/browser"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/dispatcher"
	"github.com/linkai-io/am/services/module/brute"
	"github.com/linkai-io/am/services/module/ns"
	"github.com/linkai-io/am/services/module/web"
)

var inputFile string

func init() {
	flag.StringVar(&inputFile, "input", "testdata/netflix.txt", "input file to use")
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orgID := 1
	userID := 1
	groupID := 1

	sgClient := mockScanGroupService(orgID, groupID)

	addrFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("error opening test file: %s\n", err)
	}
	addresses := amtest.AddrsFromInputFile(orgID, groupID, addrFile, nil)
	addrFile.Close()

	addrClient := mockAddressService(orgID, addresses)

	// init NS module state system & NS module
	nsstate := amtest.MockNSState()
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 2)
	nsModule := ns.New(dc, nsstate)
	if err := nsModule.Init(nil); err != nil {
		log.Fatalf("error initializing ns module: %s\n", err)
	}
	modules := make(map[am.ModuleType]am.ModuleService)
	modules[am.NSModule] = nsModule

	// init brute module state system & brute module
	brutestate := amtest.MockBruteState()
	bruteModule := brute.New(dc, brutestate)
	bruteFile, err := os.Open("testdata/10.txt")
	if err != nil {
		log.Fatalf("error opening brute sub domain file: %v\n", err)
	}

	if err := bruteModule.Init(bruteFile); err != nil {
		log.Fatalf("error initializing brute force module: %v\n", err)
	}
	modules[am.BruteModule] = bruteModule

	// init web module
	browsers := browser.NewGCDBrowserPool(5)
	if err := browsers.Init(); err != nil {
		log.Fatalf("failed initializing browsers: %v\n", err)
	}
	defer browsers.Close(ctx)

	webstate := amtest.MockWebState()
	webstorage := amtest.MockStorage()

	webModule := web.New(browsers, dc, webstate, webstorage)
	if err := webModule.Init(); err != nil {
		log.Fatalf("failed to init web module: %v\n", err)
	}
	modules[am.WebModule] = webModule

	// init dispatcher
	wg := &sync.WaitGroup{}
	disState := mockDispatcherState(wg)

	dispatcher := dispatcher.New(sgClient, addrClient, modules, disState)
	dispatcher.Init(nil)

	userContext := amtest.CreateUserContext(orgID, userID)

	// Run pipeline
	dispatcher.PushAddresses(ctx, userContext, groupID)

	wg.Wait()
}

func mockAddressService(orgID int, addresses []*am.ScanGroupAddress) *mock.AddressService {
	callCount := 0

	addrClient := &mock.AddressService{}
	addrClient.GetFn = func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (int, []*am.ScanGroupAddress, error) {
		if callCount == 0 {
			callCount++
			return orgID, addresses, nil
		}
		return orgID, nil, nil
	}
	addrClient.UpdateFn = func(ctx context.Context, userContext am.UserContext, addrs map[string]*am.ScanGroupAddress) (int, int, error) {
		log.Printf("adding %d addresses\n", len(addrs))
		return orgID, len(addrs), nil
	}
	return addrClient
}
func mockScanGroupService(orgID, groupID int) *mock.ScanGroupService {
	sgClient := &mock.ScanGroupService{}
	sgClient.GetFn = func(ctx context.Context, userContext am.UserContext, groupID int) (int, *am.ScanGroup, error) {
		scangroup := &am.ScanGroup{
			OrgID:                orgID,
			GroupID:              groupID,
			ModuleConfigurations: amtest.CreateModuleConfig(),
		}
		return orgID, scangroup, nil
	}

	return sgClient
}

func mockDispatcherState(wg *sync.WaitGroup) *mock.DispatcherState {
	ctx := context.Background()
	count := 0
	// init Dispatcher state system and DispatcherService
	disState := &mock.DispatcherState{}
	stateAddrs := make([]*am.ScanGroupAddress, 0)        // addresses stored in state
	stateHashes := make(map[string]*am.ScanGroupAddress) // hashes stored in state
	stateLock := &sync.RWMutex{}

	disState.PutAddressesFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses []*am.ScanGroupAddress) error {
		stateLock.Lock()
		defer stateLock.Unlock()
		stateAddrs = append(stateAddrs, addresses...)
		for _, addr := range addresses {
			stateHashes[addr.AddressHash] = addr
			count++
		}
		log.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!TOTAL %d, len state: %d len hashes: %d\n", count, len(stateAddrs), len(stateHashes))
		return nil
	}

	disState.PutAddressMapFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses map[string]*am.ScanGroupAddress) error {
		stateLock.Lock()
		defer stateLock.Unlock()
		for _, v := range addresses {
			stateAddrs = append(stateAddrs, v)
			stateHashes[v.AddressHash] = v
			count++
		}
		log.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!TOTAL %d, len state: %d len hashes: %d\n", count, len(stateAddrs), len(stateHashes))

		return nil
	}

	disState.FilterNewFn = func(ctx context.Context, orgID, scanGroupID int, addresses map[string]*am.ScanGroupAddress) (map[string]*am.ScanGroupAddress, error) {
		stateLock.Lock()
		defer stateLock.Unlock()
		for k := range addresses {
			if _, exist := stateHashes[k]; exist {
				delete(addresses, k)
			}
		}
		return addresses, nil
	}

	disState.PopAddressesFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[string]*am.ScanGroupAddress, error) {
		stateLock.Lock()
		defer stateLock.Unlock()
		newAddrs := make(map[string]*am.ScanGroupAddress)
		for _, addr := range stateAddrs {
			newAddrs[addr.AddressHash] = addr
			count--
		}

		// clear out addresses
		stateAddrs = make([]*am.ScanGroupAddress, 0)
		log.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!after pop, len: %d, %v\n", len(newAddrs), newAddrs)
		return newAddrs, nil
	}

	wg.Add(1)
	disState.StopFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
		wg.Done()
		return nil
	}

	go printStats(ctx, stateLock, count, stateAddrs, stateHashes)

	return disState
}

func printStats(ctx context.Context, lock *sync.RWMutex, count int, stateAddrs []*am.ScanGroupAddress, stateHashes map[string]*am.ScanGroupAddress) {
	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lock.RLock()
			log.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!TOTAL %d, len state: %d len hashes: %d\n", count, len(stateAddrs), len(stateHashes))
			lock.RUnlock()
		}
	}
}
