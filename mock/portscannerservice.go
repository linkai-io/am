package mock

import (
	"context"

	"github.com/linkai-io/am/am"
)

type PortScannerService struct {
	AddGroupFn      func(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) error
	AddGroupInvoked bool

	RemoveGroupFn      func(ctx context.Context, userContext am.UserContext, orgID, groupID int) error
	RemoveGroupInvoked bool

	AnalyzeFn      func(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, *am.PortResults, error)
	AnalyzeInvoked bool
}

func (s *PortScannerService) AddGroup(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) error {
	s.AddGroupInvoked = true
	return s.AddGroupFn(ctx, userContext, group)
}

func (s *PortScannerService) RemoveGroup(ctx context.Context, userContext am.UserContext, orgID, groupID int) error {
	s.RemoveGroupInvoked = true
	return s.RemoveGroupFn(ctx, userContext, orgID, groupID)
}

func (s *PortScannerService) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, *am.PortResults, error) {
	s.AnalyzeInvoked = true
	return s.AnalyzeFn(ctx, userContext, address)
}

