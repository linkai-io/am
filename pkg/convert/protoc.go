package convert

import (
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/protocservices/prototypes"
)

// UserContextToDomain converts from a protoc usercontext to an am.usercontext
func UserContextToDomain(in *prototypes.UserContext) am.UserContext {
	return &am.UserContextData{
		TraceID:   in.TraceID,
		OrgID:     int(in.OrgID),
		UserID:    int(in.UserID),
		Roles:     in.Roles,
		IPAddress: in.IPAddress,
	}
}

func DomainToOrganization(in *am.Organization) *prototypes.Org {
	return &prototypes.Org{
		OrgID:           int32(in.OrgID),
		OrgCID:          in.OrgCID,
		OrgName:         in.OrgName,
		OwnerEmail:      in.OwnerEmail,
		UserPoolID:      in.UserPoolID,
		IdentityPoolID:  in.IdentityPoolID,
		FirstName:       in.FirstName,
		LastName:        in.LastName,
		Phone:           in.Phone,
		Country:         in.Country,
		StatePrefecture: in.StatePrefecture,
		Street:          in.Street,
		Address1:        in.Address1,
		Address2:        in.Address2,
		City:            in.City,
		PostalCode:      in.PostalCode,
		CreationTime:    in.CreationTime,
		StatusID:        int32(in.StatusID),
		Deleted:         in.Deleted,
		SubscriptionID:  int32(in.SubscriptionID),
	}
}

func OrganizationToDomain(in *prototypes.Org) *am.Organization {
	return &am.Organization{
		OrgID:           int(in.OrgID),
		OrgCID:          in.OrgCID,
		OrgName:         in.OrgName,
		OwnerEmail:      in.OwnerEmail,
		UserPoolID:      in.UserPoolID,
		IdentityPoolID:  in.IdentityPoolID,
		FirstName:       in.FirstName,
		LastName:        in.LastName,
		Phone:           in.Phone,
		Country:         in.Country,
		StatePrefecture: in.StatePrefecture,
		Street:          in.Street,
		Address1:        in.Address1,
		Address2:        in.Address2,
		City:            in.City,
		PostalCode:      in.PostalCode,
		CreationTime:    in.CreationTime,
		StatusID:        int(in.StatusID),
		Deleted:         in.Deleted,
		SubscriptionID:  int(in.SubscriptionID),
	}
}

func OrgFilterToDomain(in *prototypes.OrgFilter) *am.OrgFilter {
	return &am.OrgFilter{
		Start:             int(in.Start),
		Limit:             int(in.Limit),
		WithDeleted:       in.WithDeleted,
		DeletedValue:      in.DeletedValue,
		WithStatus:        in.WithStatus,
		StatusValue:       in.StatusValue,
		WithSubcription:   in.WithSubcription,
		SubValue:          in.SubValue,
		SinceCreationTime: in.SinceCreationTime,
	}
}
