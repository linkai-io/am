package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/metrics/load"
	"github.com/linkai-io/am/protocservices/module/portscan"
	context "golang.org/x/net/context"
)

type PortScanModuleProtocService struct {
	portscan am.PortScannerService
	reporter *load.RateReporter
}

func New(implementation am.PortScannerService, reporter *load.RateReporter) *PortScanModuleProtocService {
	return &PortScanModuleProtocService{portscan: implementation, reporter: reporter}
}

func (s *PortScanModuleProtocService) AddGroup(ctx context.Context, in *portscan.PortScanAddGroupRequest) (*portscan.PortScanGroupAddedResponse, error) {
	s.reporter.Increment(1)
	err := s.portscan.AddGroup(ctx, convert.UserContextToDomain(in.UserContext), convert.ScanGroupToDomain(in.Group))
	s.reporter.Increment(-1)
	return &portscan.PortScanGroupAddedResponse{}, err
}

func (s *PortScanModuleProtocService) RemoveGroup(ctx context.Context, in *portscan.PortScanRemoveGroupRequest) (*portscan.PortScanGroupRemovedResponse, error) {
	s.reporter.Increment(1)
	err := s.portscan.RemoveGroup(ctx, convert.UserContextToDomain(in.UserContext), int(in.OrgID), int(in.GroupID))
	s.reporter.Increment(-1)
	return &portscan.PortScanGroupRemovedResponse{}, err
}

func (s *PortScanModuleProtocService) Analyze(ctx context.Context, in *portscan.PortScanAnalyzeRequest) (*portscan.PortScanAnalyzedResponse, error) {
	s.reporter.Increment(1)
	address, portResults, err := s.portscan.Analyze(ctx, convert.UserContextToDomain(in.UserContext), convert.AddressToDomain(in.Address))
	s.reporter.Increment(-1)
	return &portscan.PortScanAnalyzedResponse{Address: convert.DomainToAddress(address), PortResult: convert.DomainToPortResults(portResults)}, err
}
