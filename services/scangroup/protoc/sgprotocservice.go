package protoc

import (
	"errors"

	context "golang.org/x/net/context"
	"gopkg.linkai.io/v1/repos/am/am"
)

var (
	ErrOrgIDNonMatch = errors.New("error organization id's did not match")
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
	groupID, versionID, err := s.sgs.Create(ctx, in.Group.OrgID, in.RequesterUserID, groupToDomain(in.Group), groupVersionToDomain(in.Version))
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
	orgID, groupVersion, err := s.sgs.GetVersion(ctx, in.OrgID, in.RequesterUserID, in.GroupID, in.GroupVersionID)
	if err != nil {
		return nil, err
	}
	return &GroupVersionResponse{OrgID: orgID, GroupVersion: domainToGroupVersion(groupVersion)}, nil
}

func (s *SGProtocService) CreateVersion(ctx context.Context, in *NewVersionRequest) (*VersionCreatedResponse, error) {
	oid, gid, gvid, err := s.sgs.CreateVersion(ctx, in.Version.OrgID, in.RequesterUserID, groupVersionToDomain(in.Version))
	if err != nil {
		return nil, err
	}
	return &VersionCreatedResponse{OrgID: oid, GroupID: gid, GroupVersionID: gvid}, nil
}

func (s *SGProtocService) DeleteVersion(ctx context.Context, in *DeleteVersionRequest) (*VersionDeletedResponse, error) {
	oid, gid, gvid, err := s.sgs.DeleteVersion(ctx, in.OrgID, in.RequesterUserID, in.GroupID, in.GroupVersionID, in.VersionName)
	if err != nil {
		return nil, err
	}
	return &VersionDeletedResponse{OrgID: oid, GroupID: gid, GroupVersionID: gvid}, nil
}

func (s *SGProtocService) Groups(in *GroupsRequest, stream ScanGroup_GroupsServer) error {
	ctx := context.Background()

	oid, groups, err := s.sgs.Groups(ctx, in.OrgID)
	if err != nil {
		return err
	}

	for _, g := range groups {
		if oid != g.OrgID {
			return ErrOrgIDNonMatch
		}

		if err := stream.Send(&GroupResponse{Group: domainToGroup(g)}); err != nil {
			return err
		}
	}
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

func domainToGroupVersion(in *am.ScanGroupVersion) *GroupVersion {
	return &GroupVersion{
		OrgID:          in.OrgID,
		GroupID:        in.GroupID,
		GroupVersionID: in.GroupVersionID,
		VersionName:    in.VersionName,
		CreationTime:   in.CreationTime,
		CreatedBy:      in.CreatedBy,
		Configuration:  domainToModule(in.ModuleConfigurations),
		Deleted:        in.Deleted,
	}
}

func groupVersionToDomain(in *GroupVersion) *am.ScanGroupVersion {
	return &am.ScanGroupVersion{
		OrgID:                in.OrgID,
		GroupID:              in.GroupID,
		GroupVersionID:       in.GroupVersionID,
		VersionName:          in.VersionName,
		CreationTime:         in.CreationTime,
		CreatedBy:            in.CreatedBy,
		ModuleConfigurations: moduleToDomain(in.Configuration),
		Deleted:              in.Deleted,
	}
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

func domainToModule(in *am.ModuleConfiguration) *ModuleConfiguration {
	return &ModuleConfiguration{
		NSConfig:        &NSModuleConfig{Name: in.NSModule.Name},
		BruteConfig:     &BruteModuleConfig{Name: in.BruteModule.Name, CustomSubNames: in.BruteModule.CustomSubNames, MaxDepth: in.BruteModule.MaxDepth},
		PortConfig:      &PortModuleConfig{Name: in.PortModule.Name, Ports: in.PortModule.Ports},
		WebModuleConfig: &WebModuleConfig{Name: in.WebModule.Name, TakeScreenShots: in.WebModule.TakeScreenShots, MaxLinks: in.WebModule.MaxLinks, ExtractJS: in.WebModule.ExtractJS, FingerprintFrameworks: in.WebModule.FingerprintFrameworks},
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
