package protoc

import (
	"errors"

	context "golang.org/x/net/context"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/convert"
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
	oid, group, err := s.sgs.Get(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupResponse{OrgID: int32(oid), Group: convert.DomainToGroup(group)}, err
}

func (s *SGProtocService) GetByName(ctx context.Context, in *scangroup.GroupRequest) (*scangroup.GroupResponse, error) {
	oid, group, err := s.sgs.GetByName(ctx, convert.UserContextToDomain(in.UserContext), in.GroupName)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupResponse{OrgID: int32(oid), Group: convert.DomainToGroup(group)}, err
}

func (s *SGProtocService) Create(ctx context.Context, in *scangroup.NewGroupRequest) (*scangroup.GroupCreatedResponse, error) {
	orgID, groupID, err := s.sgs.Create(ctx, convert.UserContextToDomain(in.UserContext), convert.GroupToDomain(in.Group))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupCreatedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Update(ctx context.Context, in *scangroup.UpdateGroupRequest) (*scangroup.GroupUpdatedResponse, error) {
	orgID, groupID, err := s.sgs.Update(ctx, convert.UserContextToDomain(in.UserContext), convert.GroupToDomain(in.Group))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupUpdatedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Delete(ctx context.Context, in *scangroup.DeleteGroupRequest) (*scangroup.GroupDeletedResponse, error) {
	orgID, groupID, err := s.sgs.Delete(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupDeletedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Groups(in *scangroup.GroupsRequest, stream scangroup.ScanGroup_GroupsServer) error {
	oid, groups, err := s.sgs.Groups(stream.Context(), convert.UserContextToDomain(in.UserContext))
	if err != nil {
		return err
	}

	for _, g := range groups {
		if oid != g.OrgID {
			return ErrOrgIDNonMatch
		}

		if err := stream.Send(&scangroup.GroupResponse{Group: convert.DomainToGroup(g)}); err != nil {
			return err
		}
	}
	return nil
}
