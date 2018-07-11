package protoc

import (
	"errors"

	context "golang.org/x/net/context"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/protocservices/scangroup"
)

var (
	ErrOrgIDNonMatch      = errors.New("error organization id's did not match")
	ErrMissingUserContext = errors.New("error request was missing user context")
)

type SGProtocService struct {
	sgs am.ScanGroupService
}

func New(implementation am.ScanGroupService) *SGProtocService {
	return &SGProtocService{sgs: implementation}
}

func (s *SGProtocService) Get(ctx context.Context, in *scangroup.GroupRequest) (*scangroup.GroupResponse, error) {
	oid, group, err := s.sgs.Get(ctx, in.UserContext, in.GroupID)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupResponse{OrgID: oid, Group: domainToGroup(group)}, err
}

func (s *SGProtocService) GetByName(ctx context.Context, in *scangroup.GroupRequest) (*scangroup.GroupResponse, error) {
	oid, group, err := s.sgs.GetByName(ctx, in.UserContext, in.GroupName)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupResponse{OrgID: oid, Group: domainToGroup(group)}, err
}

func (s *SGProtocService) Create(ctx context.Context, in *scangroup.NewGroupRequest) (*scangroup.GroupCreatedResponse, error) {
	orgID, groupID, groupVersionID, err := s.sgs.Create(ctx, in.UserContext, groupToDomain(in.Group), groupVersionToDomain(in.Version))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupCreatedResponse{OrgID: orgID, GroupID: groupID, GroupVersionID: groupVersionID}, nil
}

func (s *SGProtocService) Delete(ctx context.Context, in *scangroup.DeleteGroupRequest) (*scangroup.GroupDeletedResponse, error) {
	orgID, groupID, err := s.sgs.Delete(ctx, in.UserContext, in.GroupID)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupDeletedResponse{OrgID: orgID, GroupID: groupID}, nil
}

func (s *SGProtocService) GetVersion(ctx context.Context, in *scangroup.GroupVersionRequest) (*scangroup.GroupVersionResponse, error) {
	orgID, groupVersion, err := s.sgs.GetVersion(ctx, in.UserContext, in.GroupID, in.GroupVersionID)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupVersionResponse{OrgID: orgID, GroupVersion: domainToGroupVersion(groupVersion)}, nil
}

func (s *SGProtocService) GetVersionByName(ctx context.Context, in *scangroup.GroupVersionRequest) (*scangroup.GroupVersionResponse, error) {
	orgID, groupVersion, err := s.sgs.GetVersionByName(ctx, in.UserContext, in.GroupID, in.GroupVersionName)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupVersionResponse{OrgID: orgID, GroupVersion: domainToGroupVersion(groupVersion)}, nil
}

func (s *SGProtocService) CreateVersion(ctx context.Context, in *scangroup.NewVersionRequest) (*scangroup.VersionCreatedResponse, error) {
	oid, gid, gvid, err := s.sgs.CreateVersion(ctx, in.UserContext, groupVersionToDomain(in.Version))
	if err != nil {
		return nil, err
	}
	return &scangroup.VersionCreatedResponse{OrgID: oid, GroupID: gid, GroupVersionID: gvid}, nil
}

func (s *SGProtocService) DeleteVersion(ctx context.Context, in *scangroup.DeleteVersionRequest) (*scangroup.VersionDeletedResponse, error) {
	oid, gid, gvid, err := s.sgs.DeleteVersion(ctx, in.UserContext, in.GroupID, in.GroupVersionID, in.VersionName)
	if err != nil {
		return nil, err
	}
	return &scangroup.VersionDeletedResponse{OrgID: oid, GroupID: gid, GroupVersionID: gvid}, nil
}

func (s *SGProtocService) Groups(in *scangroup.GroupsRequest, stream scangroup.ScanGroup_GroupsServer) error {
	oid, groups, err := s.sgs.Groups(stream.Context(), in.UserContext)
	if err != nil {
		return err
	}

	for _, g := range groups {
		if oid != g.OrgID {
			return ErrOrgIDNonMatch
		}

		if err := stream.Send(&scangroup.GroupResponse{Group: domainToGroup(g)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SGProtocService) Addresses(in *scangroup.AddressesRequest, stream scangroup.ScanGroup_AddressesServer) error {
	return nil
}

func (s *SGProtocService) AddAddresses(stream scangroup.ScanGroup_AddAddressesServer) error {
	return nil
}

func (s *SGProtocService) UpdatedAddresses(stream scangroup.ScanGroup_UpdatedAddressesServer) error {
	return nil
}

func domainToGroupVersion(in *am.ScanGroupVersion) *scangroup.GroupVersion {
	return &scangroup.GroupVersion{
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

func groupVersionToDomain(in *scangroup.GroupVersion) *am.ScanGroupVersion {
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

func addressToDomain(in *scangroup.Address) *am.ScanGroupAddress {
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
func moduleToDomain(in *scangroup.ModuleConfiguration) *am.ModuleConfiguration {
	return &am.ModuleConfiguration{
		NSModule:    &am.NSModuleConfig{Name: in.NSConfig.Name},
		BruteModule: &am.BruteModuleConfig{Name: in.BruteConfig.Name, CustomSubNames: in.BruteConfig.CustomSubNames, MaxDepth: in.BruteConfig.MaxDepth},
		PortModule:  &am.PortModuleConfig{Name: in.PortConfig.Name, Ports: in.PortConfig.Ports},
		WebModule:   &am.WebModuleConfig{Name: in.WebModuleConfig.Name, TakeScreenShots: in.WebModuleConfig.TakeScreenShots, MaxLinks: in.WebModuleConfig.MaxLinks, ExtractJS: in.WebModuleConfig.ExtractJS, FingerprintFrameworks: in.WebModuleConfig.FingerprintFrameworks},
	}
}

func domainToModule(in *am.ModuleConfiguration) *scangroup.ModuleConfiguration {
	return &scangroup.ModuleConfiguration{
		NSConfig:        &scangroup.NSModuleConfig{Name: in.NSModule.Name},
		BruteConfig:     &scangroup.BruteModuleConfig{Name: in.BruteModule.Name, CustomSubNames: in.BruteModule.CustomSubNames, MaxDepth: in.BruteModule.MaxDepth},
		PortConfig:      &scangroup.PortModuleConfig{Name: in.PortModule.Name, Ports: in.PortModule.Ports},
		WebModuleConfig: &scangroup.WebModuleConfig{Name: in.WebModule.Name, TakeScreenShots: in.WebModule.TakeScreenShots, MaxLinks: in.WebModule.MaxLinks, ExtractJS: in.WebModule.ExtractJS, FingerprintFrameworks: in.WebModule.FingerprintFrameworks},
	}
}

// groupToDomain convert protoc group to domain type ScanGroup
func groupToDomain(in *scangroup.Group) *am.ScanGroup {
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
func domainToGroup(in *am.ScanGroup) *scangroup.Group {
	return &scangroup.Group{
		OrgID:         in.OrgID,
		GroupID:       in.GroupID,
		GroupName:     in.GroupName,
		CreationTime:  in.CreationTime,
		CreatedBy:     in.CreatedBy,
		OriginalInput: in.OriginalInput,
		Deleted:       in.Deleted,
	}
}
