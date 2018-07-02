package protoc

import (
	context "golang.org/x/net/context"
	"gopkg.linkai.io/v1/repos/am/am"
)

type SGProtocService struct {
	sgs am.ScanGroupService
}

func New(implementation am.ScanGroupService) *SGProtocService {
	return &SGProtocService{sgs: implementation}
}

func (s *SGProtocService) Get(ctx context.Context, in *GroupRequest) (*GroupResponse, error) {
	return &GroupResponse{}, nil
}

func (s *SGProtocService) Create(ctx context.Context, in *NewGroupRequest) (*VersionCreatedResponse, error) {
	groupID, versionID, err := s.sgs.Create(ctx, in.Group.OrgID, in.RequesterUserID, groupToDomain(in.Group))
	if err != nil {
		return nil, err
	}
	return &VersionCreatedResponse{GroupID: groupID, GroupVersionID: versionID}, nil
}

func (s *SGProtocService) Delete(ctx context.Context, in *DeleteGroupRequest) (*GroupDeletedResponse, error) {
	orgID, groupID, err := s.sgs.Delete(ctx, in.OrgID, in.RequesterUserID, in.GroupID)
	if err != nil {
		return nil, err
	}
	return &GroupDeletedResponse{OrgID: orgID, GroupID: groupID}, nil
}

func (s *SGProtocService) GetVersion(ctx context.Context, in *GroupVersionRequest) (*GroupVersionResponse, error) {
	return &GroupVersionResponse{}, nil
}

func (s *SGProtocService) CreateVersion(ctx context.Context, in *NewVersionRequest) (*VersionCreatedResponse, error) {
	return &VersionCreatedResponse{GroupID: 1, GroupVersionID: 1}, nil
}

func (s *SGProtocService) DeleteVersion(ctx context.Context, in *DeleteVersionRequest) (*VersionDeletedResponse, error) {
	return &VersionDeletedResponse{GroupID: 1, GroupVersionID: 1}, nil
}

func (s *SGProtocService) Groups(in *GroupsRequest, stream ScanGroup_GroupsServer) error {
	return nil
}

func (s *SGProtocService) Addresses(in *AddressesRequest, stream ScanGroup_AddressesServer) error {
	return nil
}

func (s *SGProtocService) AddAddresses(stream ScanGroup_AddAddressesServer) error {
	return nil
}

func (s *SGProtocService) UpdatedAddresses(stream ScanGroup_UpdatedAddressesServer) error {
	return nil
}

func addressToDomain(in *Address) *am.ScanGroupAddress {
	return &am.ScanGroupAddress{
		OrgID:     in.OrgID,
		AddressID: in.AddressID,
		GroupID:   in.GroupID,
		Address:   in.Address,
		Settings:  moduleToDomain(in.Settings),
		AddedTime: in.AddedTime,
		AddedBy:   in.AddedBy,
		Ignored:   in.Ignored,
	}
}

// moduleToDomain converts protoc ModuleConfiguration to am.ModuleConfiguration
func moduleToDomain(in *ModuleConfiguration) *am.ModuleConfiguration {
	return &am.ModuleConfiguration{
		NSModule:    &am.NSModuleConfig{Name: in.NSConfig.Name},
		BruteModule: &am.BruteModuleConfig{Name: in.BruteConfig.Name, CustomSubNames: in.BruteConfig.CustomSubNames, MaxDepth: in.BruteConfig.MaxDepth},
		PortModule:  &am.PortModuleConfig{Name: in.PortConfig.Name, Ports: in.PortConfig.Ports},
		WebModule:   &am.WebModuleConfig{Name: in.WebModuleConfig.Name, TakeScreenShots: in.WebModuleConfig.TakeScreenShots, MaxLinks: in.WebModuleConfig.MaxLinks, ExtractJS: in.WebModuleConfig.ExtractJS, FingerprintFrameworks: in.WebModuleConfig.FingerprintFrameworks},
	}
}

// groupToDomain convert protoc group to domain type ScanGroup
func groupToDomain(in *Group) *am.ScanGroup {
	return &am.ScanGroup{
		OrgID:         in.OrgID,
		GroupID:       in.GroupID,
		GroupName:     in.GroupName,
		CreationTime:  in.CreationTime,
		CreatedBy:     in.CreatedBy,
		OriginalInput: in.OriginalInput,
		Deleted:       in.Deleted,
	}
}

// domainToGroup convert domain type SdcanGroup to protoc Group
func domainToGroup(in *am.ScanGroup) *Group {
	return &Group{
		OrgID:         in.OrgID,
		GroupID:       in.GroupID,
		GroupName:     in.GroupName,
		CreationTime:  in.CreationTime,
		CreatedBy:     in.CreatedBy,
		OriginalInput: in.OriginalInput,
		Deleted:       in.Deleted,
	}
}
