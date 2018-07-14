package protoc

import (
	"errors"

	"gopkg.linkai.io/v1/repos/am/protocservices/prototypes"

	context "golang.org/x/net/context"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/protocservices/scangroup"
)

var (
	ErrOrgIDNonMatch      = errors.New("error organization id's did not match")
	ErrMissingUserContext = errors.New("error request was missing user context")
)

type SGProtocService struct {
	sgs              am.ScanGroupService
	MaxAddressStream int32
}

func New(implementation am.ScanGroupService) *SGProtocService {
	return &SGProtocService{sgs: implementation, MaxAddressStream: 200}
}

func (s *SGProtocService) Get(ctx context.Context, in *scangroup.GroupRequest) (*scangroup.GroupResponse, error) {
	oid, group, err := s.sgs.Get(ctx, userContextToDomain(in.UserContext), int(in.GroupID))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupResponse{OrgID: int32(oid), Group: domainToGroup(group)}, err
}

func (s *SGProtocService) GetByName(ctx context.Context, in *scangroup.GroupRequest) (*scangroup.GroupResponse, error) {
	oid, group, err := s.sgs.GetByName(ctx, userContextToDomain(in.UserContext), in.GroupName)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupResponse{OrgID: int32(oid), Group: domainToGroup(group)}, err
}

func (s *SGProtocService) Create(ctx context.Context, in *scangroup.NewGroupRequest) (*scangroup.GroupCreatedResponse, error) {
	orgID, groupID, err := s.sgs.Create(ctx, userContextToDomain(in.UserContext), groupToDomain(in.Group))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupCreatedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Update(ctx context.Context, in *scangroup.UpdateGroupRequest) (*scangroup.GroupUpdatedResponse, error) {
	orgID, groupID, err := s.sgs.Update(ctx, userContextToDomain(in.UserContext), groupToDomain(in.Group))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupUpdatedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Delete(ctx context.Context, in *scangroup.DeleteGroupRequest) (*scangroup.GroupDeletedResponse, error) {
	orgID, groupID, err := s.sgs.Delete(ctx, userContextToDomain(in.UserContext), int(in.GroupID))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupDeletedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Groups(in *scangroup.GroupsRequest, stream scangroup.ScanGroup_GroupsServer) error {
	oid, groups, err := s.sgs.Groups(stream.Context(), userContextToDomain(in.UserContext))
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
	filter := &am.ScanGroupAddressFilter{GroupID: int(in.GroupID), Start: int(in.Start), Limit: int(in.Limit), Deleted: in.Deleted, Ignored: in.Ignored}
	oid, addresses, err := s.sgs.Addresses(stream.Context(), userContextToDomain(in.UserContext), filter)
	if err != nil {
		return err
	}

	for _, a := range addresses {
		if oid != a.OrgID {
			return ErrOrgIDNonMatch
		}

		if err := stream.Send(&scangroup.AddressesResponse{Addresses: domainToAddress(a)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SGProtocService) AddressCount(ctx context.Context, in *scangroup.CountAddressesRequest) (*scangroup.CountAddressesResponse, error) {
	oid, count, err := s.sgs.AddressCount(ctx, userContextToDomain(in.UserContext), int(in.GroupID))
	if err != nil {
		return nil, err
	}
	return &scangroup.CountAddressesResponse{OrgID: int32(oid), GroupID: in.GroupID, Count: int32(count)}, nil
}

func (s *SGProtocService) AddAddresses(ctx context.Context, in *scangroup.AddAddressRequest) (*scangroup.AddAddressesResponse, error) {

	header := &am.ScanGroupAddressHeader{GroupID: int(in.GroupID), AddedBy: in.AddedBy, Ignored: in.Ignored}

	oid, err := s.sgs.AddAddresses(ctx, userContextToDomain(in.UserContext), header, in.Address)
	if err != nil {
		return nil, err
	}
	return &scangroup.AddAddressesResponse{OrgID: int32(oid)}, nil
}

func (s *SGProtocService) UpdatedAddresses(ctx context.Context, in *scangroup.UpdateAddressRequest) (*scangroup.UpdateAddressesResponse, error) {
	var oid int
	var err error

	if in.Delete {
		oid, err = s.sgs.DeleteAddresses(ctx, userContextToDomain(in.UserContext), int(in.GroupID), in.AddressMap)
	} else {
		oid, err = s.sgs.IgnoreAddresses(ctx, userContextToDomain(in.UserContext), int(in.GroupID), in.AddressMap)
	}
	return &scangroup.UpdateAddressesResponse{OrgID: int32(oid)}, err
}

func userContextToDomain(in *prototypes.UserContext) am.UserContext {
	return &am.UserContextData{
		TraceID:   in.TraceID,
		OrgID:     int(in.OrgID),
		UserID:    int(in.UserID),
		Roles:     in.Roles,
		IPAddress: in.IPAddress,
	}
}

func addressToDomain(in *scangroup.Address) *am.ScanGroupAddress {
	return &am.ScanGroupAddress{
		OrgID:     int(in.OrgID),
		AddressID: in.AddressID,
		GroupID:   int(in.GroupID),
		Address:   in.Address,
		//Settings:  moduleToDomain(in.Settings),
		AddedTime: in.AddedTime,
		AddedBy:   in.AddedBy,
		Ignored:   in.Ignored,
		Deleted:   in.Deleted,
	}
}

func domainToAddress(in *am.ScanGroupAddress) *scangroup.Address {
	return &scangroup.Address{
		OrgID:     int32(in.OrgID),
		AddressID: in.AddressID,
		GroupID:   int32(in.GroupID),
		Address:   in.Address,
		AddedTime: in.AddedTime,
		AddedBy:   in.AddedBy,
		Ignored:   in.Ignored,
		Deleted:   in.Deleted,
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
		OrgID:                int(in.OrgID),
		GroupID:              int(in.GroupID),
		GroupName:            in.GroupName,
		CreationTime:         in.CreationTime,
		CreatedBy:            int(in.CreatedBy),
		OriginalInput:        in.OriginalInput,
		ModuleConfigurations: moduleToDomain(in.ModuleConfiguration),
		Deleted:              in.Deleted,
	}
}

// domainToGroup convert domain type SdcanGroup to protoc Group
func domainToGroup(in *am.ScanGroup) *scangroup.Group {
	return &scangroup.Group{
		OrgID:               int32(in.OrgID),
		GroupID:             int32(in.GroupID),
		GroupName:           in.GroupName,
		CreationTime:        in.CreationTime,
		CreatedBy:           int32(in.CreatedBy),
		OriginalInput:       in.OriginalInput,
		ModuleConfiguration: domainToModule(in.ModuleConfigurations),
		Deleted:             in.Deleted,
	}
}
