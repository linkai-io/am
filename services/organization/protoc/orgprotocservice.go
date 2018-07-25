package protoc

import (
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/convert"
	"gopkg.linkai.io/v1/repos/am/protocservices/organization"

	context "golang.org/x/net/context"
)

type OrgProtocService struct {
	orgservice am.OrganizationService
}

func New(implementation am.OrganizationService) *OrgProtocService {
	return &OrgProtocService{orgservice: implementation}
}

func (o *OrgProtocService) Get(ctx context.Context, in *organization.OrgRequest) (*organization.OrgResponse, error) {
	var err error
	var org *am.Organization
	var oid int

	switch in.By {
	case organization.OrgRequest_ORGNAME:
		oid, org, err = o.orgservice.Get(ctx, convert.UserContextToDomain(in.UserContext), in.OrgName)
	case organization.OrgRequest_ORGCID:
		oid, org, err = o.orgservice.GetByCID(ctx, convert.UserContextToDomain(in.UserContext), in.OrgCID)
	case organization.OrgRequest_ORGID:
		oid, org, err = o.orgservice.GetByID(ctx, convert.UserContextToDomain(in.UserContext), int(in.OrgID))
	}
	return &organization.OrgResponse{OrgID: int32(oid), Org: convert.DomainToOrganization(org)}, err
}

func (o *OrgProtocService) List(in *organization.OrgListRequest, stream organization.Organization_ListServer) error {
	orgs, err := o.orgservice.List(stream.Context(), convert.UserContextToDomain(in.UserContext), convert.OrgFilterToDomain(in.OrgFilter))
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

func (o *OrgProtocService) Create(ctx context.Context, in *organization.CreateOrgRequest) (*organization.OrgCreatedResponse, error) {
	orgID, userID, orgCID, userCID, err := o.orgservice.Create(ctx, convert.UserContextToDomain(in.UserContext), convert.OrganizationToDomain(in.Org))
	if err != nil {
		return nil, err
	}
	return &organization.OrgCreatedResponse{OrgID: int32(orgID), UserID: int32(userID), OrgCID: orgCID, UserCID: userCID}, nil
}

func (o *OrgProtocService) Update(ctx context.Context, in *organization.UpdateOrgRequest) (*organization.OrgUpdatedResponse, error) {
	oid, err := o.orgservice.Update(ctx, convert.UserContextToDomain(in.UserContext), convert.OrganizationToDomain(in.Org))
	if err != nil {
		return nil, err
	}
	// TODO: Fix get orgid
	return &organization.OrgUpdatedResponse{OrgID: int32(oid)}, nil
}

func (o *OrgProtocService) Delete(ctx context.Context, in *organization.DeleteOrgRequest) (*organization.OrgDeletedResponse, error) {
	oid, err := o.orgservice.Delete(ctx, convert.UserContextToDomain(in.UserContext), int(in.OrgID))
	if err != nil {
		return nil, err
	}
	return &organization.OrgDeletedResponse{OrgID: int32(oid)}, nil
}
