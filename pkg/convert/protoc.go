package convert

import (
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/protocservices/prototypes"
)

// DomainToUser convert domain user type to protobuf user type
func DomainToUser(in *am.User) *prototypes.User {
	return &prototypes.User{
		OrgID:        int32(in.OrgID),
		OrgCID:       in.OrgCID,
		UserCID:      in.UserCID,
		UserID:       int32(in.UserID),
		UserEmail:    in.UserEmail,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		StatusID:     int32(in.StatusID),
		CreationTime: in.CreationTime,
		Deleted:      in.Deleted,
	}
}

// UserToDomain convert protobuf user type to domain user type
func UserToDomain(in *prototypes.User) *am.User {
	return &am.User{
		OrgID:        int(in.OrgID),
		OrgCID:       in.OrgCID,
		UserCID:      in.UserCID,
		UserID:       int(in.UserID),
		UserEmail:    in.UserEmail,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		StatusID:     int(in.StatusID),
		CreationTime: in.CreationTime,
		Deleted:      in.Deleted,
	}
}

func DomainToUserFilter(in *am.UserFilter) *prototypes.UserFilter {
	return &prototypes.UserFilter{
		Start:             int32(in.Start),
		Limit:             int32(in.Limit),
		OrgID:             int32(in.OrgID),
		WithStatus:        in.WithStatus,
		StatusValue:       int32(in.StatusValue),
		WithDeleted:       in.WithDeleted,
		DeletedValue:      in.DeletedValue,
		SinceCreationTime: in.SinceCreationTime,
	}
}

func UserFilterToDomain(in *prototypes.UserFilter) *am.UserFilter {
	return &am.UserFilter{
		Start:             int(in.Start),
		Limit:             int(in.Limit),
		OrgID:             int(in.OrgID),
		WithStatus:        in.WithStatus,
		StatusValue:       int(in.StatusValue),
		WithDeleted:       in.WithDeleted,
		DeletedValue:      in.DeletedValue,
		SinceCreationTime: in.SinceCreationTime,
	}
}

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

// DomainToUserContext converts the domain usercontext to protobuf usercontext
func DomainToUserContext(in am.UserContext) *prototypes.UserContext {
	return &prototypes.UserContext{
		TraceID:   in.GetTraceID(),
		OrgID:     int32(in.GetOrgID()),
		UserID:    int32(in.GetUserID()),
		Roles:     in.GetRoles(),
		IPAddress: in.GetIPAddress(),
	}
}

// DomainToOrganization converts the domain organization to protobuf organization
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

// OrganizationToDomain converts the protobuf organization to domain organization
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

func DomainToOrgFilter(in *am.OrgFilter) *prototypes.OrgFilter {
	return &prototypes.OrgFilter{
		Start:             int32(in.Start),
		Limit:             int32(in.Limit),
		WithDeleted:       in.WithDeleted,
		DeletedValue:      in.DeletedValue,
		WithStatus:        in.WithStatus,
		StatusValue:       in.StatusValue,
		WithSubcription:   in.WithSubcription,
		SubValue:          in.SubValue,
		SinceCreationTime: in.SinceCreationTime,
	}
}

// OrgFilterToDomain convert org filter protobuf to orgfilter domain
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
