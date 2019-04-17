package amtest

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/filestorage"
	"github.com/linkai-io/am/pkg/state"
	"github.com/linkai-io/am/pkg/webtech"
)

func CreateUserContext(orgID, userID int) *mock.UserContext {
	userContext := &mock.UserContext{}
	userContext.GetOrgIDFn = func() int {
		return orgID
	}

	userContext.GetUserIDFn = func() int {
		return userID
	}

	userContext.GetOrgCIDFn = func() string {
		return "someorgcid"
	}

	userContext.GetUserCIDFn = func() string {
		return "someusercid"
	}

	userContext.GetSubscriptionIDFn = func() int32 {
		return 1000
	}

	return userContext
}

func MockAddressService(orgID int, addresses []*am.ScanGroupAddress) *mock.AddressService {
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
func MockScanGroupService(orgID, groupID int) *mock.ScanGroupService {
	sgClient := &mock.ScanGroupService{}
	sgClient.GetFn = func(ctx context.Context, userContext am.UserContext, groupID int) (int, *am.ScanGroup, error) {
		scangroup := &am.ScanGroup{
			OrgID:                orgID,
			GroupID:              groupID,
			ModuleConfigurations: CreateModuleConfig(),
		}
		return orgID, scangroup, nil
	}

	sgClient.UpdateStatsFn = func(ctx context.Context, userContext am.UserContext, stats *am.GroupStats) (int, error) {
		return orgID, nil
	}

	return sgClient
}

func MockAuthorizer() *mock.Authorizer {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}
	return auth
}

func MockRoleManager() *mock.RoleManager {
	roleManager := &mock.RoleManager{}
	roleManager.CreateRoleFn = func(role *am.Role) (string, error) {
		return "id", nil
	}
	roleManager.AddMembersFn = func(orgID int, roleID string, members []int) error {
		return nil
	}
	return roleManager
}

func MockEmptyAuthorizer() *mock.Authorizer {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}
	return auth
}

func MockEventService() *mock.EventService {
	mockEvent := &mock.EventService{}
	mockEvent.InitFn = func(config []byte) error {
		return nil
	}

	mockEvent.AddFn = func(ctx context.Context, userContext am.UserContext, events []*am.Event) error {
		return nil
	}

	mockEvent.NotifyCompleteFn = func(ctx context.Context, userContext am.UserContext, startTime int64, groupID int) error {
		return nil
	}
	return mockEvent
}

func MockStorage() *mock.Storage {
	mockStorage := &mock.Storage{}
	mockStorage.InitFn = func() error {
		return nil
	}

	mockStorage.WriteFn = func(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, data []byte) (string, string, error) {
		if data == nil || len(data) == 0 {
			return "", "", nil
		}

		hashName := convert.HashData(data)
		fileName := filestorage.PathFromData(address, hashName)
		if fileName == "null" {
			return "", "", nil
		}
		return hashName, fileName, nil
	}
	return mockStorage
}

func MockBruteState() *mock.BruteState {
	mockState := &mock.BruteState{}
	mockState.SubscribeFn = func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
		return nil
	}

	mockState.GetGroupFn = func(ctx context.Context, orgID int, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
		return nil, errors.New("group not found")
	}

	mockState.DoBruteETLDFn = func(ctx context.Context, orgID int, scanGroupID int, expireSeconds, maxAllowed int, etld string) (int, bool, error) {
		return 1, true, nil
	}

	bruteHosts := make(map[string]bool)
	mutateHosts := make(map[string]bool)
	mockState.DoBruteDomainFn = func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
		if _, ok := bruteHosts[zone]; !ok {
			bruteHosts[zone] = true
			return true, nil
		}
		return false, nil
	}

	mockState.DoMutateDomainFn = func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
		if _, ok := mutateHosts[zone]; !ok {
			mutateHosts[zone] = true
			return true, nil
		}
		return false, nil
	}
	return mockState
}

func MockWebState() *mock.WebState {
	mockState := &mock.WebState{}
	mockState.SubscribeFn = func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
		return nil
	}

	mockState.GetGroupFn = func(ctx context.Context, orgID int, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
		return nil, errors.New("group not found")
	}

	webHosts := make(map[string]bool)
	mockState.DoWebDomainFn = func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
		if _, ok := webHosts[zone]; !ok {
			webHosts[zone] = true
			return true, nil
		}
		return false, nil
	}

	return mockState
}

func MockNSState() *mock.NSState {
	mockState := &mock.NSState{}
	mockState.SubscribeFn = func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
		return nil
	}

	mockState.GetGroupFn = func(ctx context.Context, orgID int, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
		return nil, errors.New("group not found")
	}

	hosts := make(map[string]bool)
	mockState.DoNSRecordsFn = func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
		if _, ok := hosts[zone]; !ok {
			hosts[zone] = true
			return true, nil
		}
		return false, nil
	}
	return mockState
}

func MockBigDataState() *mock.BigDataState {
	mockState := &mock.BigDataState{}
	mockState.SubscribeFn = func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
		return nil
	}

	mockState.GetGroupFn = func(ctx context.Context, orgID int, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
		return nil, errors.New("group not found")
	}

	hosts := make(map[string]bool)
	mockState.DoCTDomainFn = func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
		if _, ok := hosts[zone]; !ok {
			hosts[zone] = true
			return true, nil
		}
		return false, nil
	}
	return mockState
}

func MockWebDetector() *mock.Detector {
	mockDetector := &mock.Detector{}
	mockDetector.InitFn = func(config []byte) error {
		return nil
	}

	mockDetector.JSFn = func(jsObjects []*webtech.JSObject) map[string][]*webtech.Match {
		return make(map[string][]*webtech.Match, 0)
	}

	mockDetector.HeadersFn = func(headers map[string]string) map[string][]*webtech.Match {
		return make(map[string][]*webtech.Match, 0)
	}

	mockDetector.DOMFn = func(dom string) map[string][]*webtech.Match {
		return make(map[string][]*webtech.Match, 0)
	}

	mockDetector.JSToInjectFn = func() string {
		return ""
	}

	mockDetector.JSResultsToObjectsFn = func(in interface{}) []*webtech.JSObject {
		return make([]*webtech.JSObject, 0)
	}

	mockDetector.MergeMatchesFn = func(results []map[string][]*webtech.Match) map[string]*am.WebTech {
		return make(map[string]*am.WebTech, 0)
	}

	return mockDetector
}

func MockDispatcherState(wg *sync.WaitGroup, groups []*am.ScanGroup) *mock.DispatcherState {
	count := 0
	// init Dispatcher state system and DispatcherService
	disState := &mock.DispatcherState{}
	stateAddrs := make([]*am.ScanGroupAddress, 0)        // addresses stored in state
	stateHashes := make(map[string]*am.ScanGroupAddress) // hashes stored in state

	stateGroups := groups
	stateLock := &sync.RWMutex{}

	disState.GetGroupFn = func(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
		stateLock.Lock()
		defer stateLock.Unlock()
		for _, g := range stateGroups {
			log.Printf("looking for %d have %d\n", scanGroupID, g.GroupID)
			if g.GroupID == scanGroupID {
				return g, nil
			}
		}
		return nil, nil
	}

	disState.SubscribeFn = func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
		return nil
	}

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

	disState.StopFn = func(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
		wg.Done()
		return nil
	}

	return disState
}
