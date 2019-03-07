package amtest

import (
	"context"
	"errors"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/filestorage"
	"github.com/linkai-io/am/pkg/state"
	"github.com/linkai-io/am/pkg/webtech"
)

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
