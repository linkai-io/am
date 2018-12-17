package protoc

import (
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/organization"

	context "golang.org/x/net/context"
)

type OrgProtocService struct {
	orgservice am.OrganizationService
	reporter   *load.RateReporter
}

func New(implementation am.OrganizationService, reporter *load.RateReporter) *OrgProtocService {
	return &OrgProtocService{orgservice: implementation, reporter: reporter}
}

func (s *OrgProtocService) Get(ctx context.Context, in *organization.OrgRequest) (*organization.OrgResponse, error) {
	var err error
	var org *am.Organization
	var oid int
	s.reporter.Increment(1)
	switch in.By {
	case organization.OrgRequest_ORGNAME:
		oid, org, err = s.orgservice.Get(ctx, convert.UserContextToDomain(in.UserContext), in.OrgName)
	case organization.OrgRequest_ORGCID:
		oid, org, err = s.orgservice.GetByCID(ctx, convert.UserContextToDomain(in.UserContext), in.OrgCID)
	case organization.OrgRequest_ORGID:
		oid, org, err = s.orgservice.GetByID(ctx, convert.UserContextToDomain(in.UserContext), int(in.OrgID))
	}
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &organization.OrgResponse{OrgID: int32(oid), Org: convert.DomainToOrganization(org)}, err
}

func (s *OrgProtocService) List(in *organization.OrgListRequest, stream organization.Organization_ListServer) error {
	s.reporter.Increment(1)
	defer s.reporter.Increment(-1)
	orgs, err := s.orgservice.List(stream.Context(), convert.UserContextToDomain(in.UserContext), convert.OrgFilterToDomain(in.OrgFilter))
	if err != nil {
		return err
	}

	for _, org := range orgs {
		if err := stream.Send(&organization.OrgListResponse{Org: convert.DomainToOrganization(org)}); err != nil {
			return err
		}
	}

	return nil
}

func (s *OrgProtocService) Create(ctx context.Context, in *organization.CreateOrgRequest) (*organization.OrgCreatedResponse, error) {
	s.reporter.Increment(1)
	orgID, userID, orgCID, userCID, err := s.orgservice.Create(ctx, convert.UserContextToDomain(in.UserContext), convert.OrganizationToDomain(in.Org), in.UserCID)
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &organization.OrgCreatedResponse{OrgID: int32(orgID), UserID: int32(userID), OrgCID: orgCID, UserCID: userCID}, nil
}

func (s *OrgProtocService) Update(ctx context.Context, in *organization.UpdateOrgRequest) (*organization.OrgUpdatedResponse, error) {
	s.reporter.Increment(1)
	oid, err := s.orgservice.Update(ctx, convert.UserContextToDomain(in.UserContext), convert.OrganizationToDomain(in.Org))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &organization.OrgUpdatedResponse{OrgID: int32(oid)}, nil
}

func (s *OrgProtocService) Delete(ctx context.Context, in *organization.DeleteOrgRequest) (*organization.OrgDeletedResponse, error) {
	s.reporter.Increment(1)
	oid, err := s.orgservice.Delete(ctx, convert.UserContextToDomain(in.UserContext), int(in.OrgID))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &organization.OrgDeletedResponse{OrgID: int32(oid)}, nil
}
