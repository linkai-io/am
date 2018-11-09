package protoc

import (
	"errors"

	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/scangroup"
	context "golang.org/x/net/context"
)

var (
	ErrOrgIDNonMatch      = errors.New("error organization id's did not match")
	ErrMissingUserContext = errors.New("error request was missing user context")
)

type SGProtocService struct {
	sgs      am.ScanGroupService
	reporter *load.RateReporter
}

func New(implementation am.ScanGroupService, reporter *load.RateReporter) *SGProtocService {
	return &SGProtocService{sgs: implementation, reporter: reporter}
}

func (s *SGProtocService) Get(ctx context.Context, in *scangroup.GroupRequest) (*scangroup.GroupResponse, error) {
	var oid int
	var group *am.ScanGroup
	var err error
	s.reporter.Increment(1)
	switch in.By {
	case scangroup.GroupRequest_GROUPID:
		oid, group, err = s.sgs.Get(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	case scangroup.GroupRequest_GROUPNAME:
		oid, group, err = s.sgs.GetByName(ctx, convert.UserContextToDomain(in.UserContext), in.GroupName)
	}
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupResponse{OrgID: int32(oid), Group: convert.DomainToScanGroup(group)}, err
}

func (s *SGProtocService) Create(ctx context.Context, in *scangroup.NewGroupRequest) (*scangroup.GroupCreatedResponse, error) {
	s.reporter.Increment(1)
	orgID, groupID, err := s.sgs.Create(ctx, convert.UserContextToDomain(in.UserContext), convert.ScanGroupToDomain(in.Group))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupCreatedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Update(ctx context.Context, in *scangroup.UpdateGroupRequest) (*scangroup.GroupUpdatedResponse, error) {
	s.reporter.Increment(1)
	orgID, groupID, err := s.sgs.Update(ctx, convert.UserContextToDomain(in.UserContext), convert.ScanGroupToDomain(in.Group))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupUpdatedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Delete(ctx context.Context, in *scangroup.DeleteGroupRequest) (*scangroup.GroupDeletedResponse, error) {
	s.reporter.Increment(1)
	orgID, groupID, err := s.sgs.Delete(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupDeletedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) AllGroups(in *scangroup.AllGroupsRequest, stream scangroup.ScanGroup_AllGroupsServer) error {
	s.reporter.Increment(1)
	defer s.reporter.Increment(-1)
	groups, err := s.sgs.AllGroups(stream.Context(), convert.UserContextToDomain(in.UserContext), convert.ScanGroupFilterToDomain(in.Filter))
	if err != nil {
		return err
	}

	for _, g := range groups {
		if err := stream.Send(&scangroup.AllGroupsResponse{Group: convert.DomainToScanGroup(g)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SGProtocService) Groups(in *scangroup.GroupsRequest, stream scangroup.ScanGroup_GroupsServer) error {
	s.reporter.Increment(1)
	defer s.reporter.Increment(-1)
	oid, groups, err := s.sgs.Groups(stream.Context(), convert.UserContextToDomain(in.UserContext))
	if err != nil {
		return err
	}

	for _, g := range groups {
		if oid != g.OrgID {
			return ErrOrgIDNonMatch
		}

		if err := stream.Send(&scangroup.GroupResponse{Group: convert.DomainToScanGroup(g)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SGProtocService) Pause(ctx context.Context, in *scangroup.PauseGroupRequest) (*scangroup.GroupPausedResponse, error) {
	s.reporter.Increment(1)
	orgID, groupID, err := s.sgs.Pause(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupPausedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}

func (s *SGProtocService) Resume(ctx context.Context, in *scangroup.ResumeGroupRequest) (*scangroup.GroupResumedResponse, error) {
	s.reporter.Increment(1)
	orgID, groupID, err := s.sgs.Resume(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &scangroup.GroupResumedResponse{OrgID: int32(orgID), GroupID: int32(groupID)}, nil
}
