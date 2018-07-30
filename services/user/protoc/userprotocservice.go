package protoc

import (
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/convert"
	"gopkg.linkai.io/v1/repos/am/protocservices/user"

	context "golang.org/x/net/context"
)

type UserProtocService struct {
	userservice am.UserService
}

func New(implementation am.UserService) *UserProtocService {
	return &UserProtocService{userservice: implementation}
}

func (u *UserProtocService) Get(ctx context.Context, in *user.UserRequest) (*user.UserResponse, error) {
	var err error
	var amuser *am.User
	var oid int

	switch in.By {
	case user.UserRequest_USERID:
		oid, amuser, err = u.userservice.Get(ctx, convert.UserContextToDomain(in.UserContext), int(in.UserID))
	case user.UserRequest_USERCID:
		oid, amuser, err = u.userservice.GetByCID(ctx, convert.UserContextToDomain(in.UserContext), in.UserCID)
	}
	return &user.UserResponse{OrgID: int32(oid), User: convert.DomainToUser(amuser)}, err
}

func (u *UserProtocService) List(in *user.UserListRequest, stream user.UserService_ListServer) error {
	oid, users, err := u.userservice.List(stream.Context(), convert.UserContextToDomain(in.UserContext), convert.UserFilterToDomain(in.UserFilter))
	if err != nil {
		return err
	}

	for _, amuser := range users {
		if err := stream.Send(&user.UserListResponse{OrgID: int32(oid), User: convert.DomainToUser(amuser)}); err != nil {
			return err
		}
	}
	return nil
}

func (u *UserProtocService) Create(ctx context.Context, in *user.CreateUserRequest) (*user.UserCreatedResponse, error) {
	oid, uid, userCID, err := u.userservice.Create(ctx, convert.UserContextToDomain(in.UserContext), convert.UserToDomain(in.User))
	if err != nil {
		return nil, err
	}
	return &user.UserCreatedResponse{OrgID: int32(oid), UserID: int32(uid), UserCID: userCID}, nil
}

func (u *UserProtocService) Update(ctx context.Context, in *user.UpdateUserRequest) (*user.UserUpdatedResponse, error) {
	oid, uid, err := u.userservice.Update(ctx, convert.UserContextToDomain(in.UserContext), convert.UserToDomain(in.User), int(in.UserID))
	if err != nil {
		return nil, err
	}
	return &user.UserUpdatedResponse{OrgID: int32(oid), UserID: int32(uid)}, nil
}

func (u *UserProtocService) Delete(ctx context.Context, in *user.DeleteUserRequest) (*user.UserDeletedResponse, error) {
	oid, err := u.userservice.Delete(ctx, convert.UserContextToDomain(in.UserContext), int(in.UserID))
	if err != nil {
		return nil, err
	}
	return &user.UserDeletedResponse{OrgID: int32(oid)}, nil
}
