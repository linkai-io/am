package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/metrics/load"
	"github.com/linkai-io/am/protocservices/user"

	context "golang.org/x/net/context"
)

type UserProtocService struct {
	userservice am.UserService
	reporter    *load.RateReporter
}

func New(implementation am.UserService, reporter *load.RateReporter) *UserProtocService {
	return &UserProtocService{userservice: implementation, reporter: reporter}
}

func (s *UserProtocService) Get(ctx context.Context, in *user.UserRequest) (*user.UserResponse, error) {
	var err error
	var amuser *am.User
	var oid int

	s.reporter.Increment(1)
	switch in.By {
	case user.UserRequest_USEREMAIL:
		oid, amuser, err = s.userservice.Get(ctx, convert.UserContextToDomain(in.UserContext), in.UserEmail)
	case user.UserRequest_USERWITHORGID:
		oid, amuser, err = s.userservice.GetWithOrgID(ctx, convert.UserContextToDomain(in.UserContext), int(in.OrgID), in.UserCID)
	case user.UserRequest_USERID:
		oid, amuser, err = s.userservice.GetByID(ctx, convert.UserContextToDomain(in.UserContext), int(in.UserID))
	case user.UserRequest_USERCID:
		oid, amuser, err = s.userservice.GetByCID(ctx, convert.UserContextToDomain(in.UserContext), in.UserCID)
	}
	s.reporter.Increment(-1)
	return &user.UserResponse{OrgID: int32(oid), User: convert.DomainToUser(amuser)}, err
}

func (s *UserProtocService) List(in *user.UserListRequest, stream user.UserService_ListServer) error {
	s.reporter.Increment(1)
	defer s.reporter.Increment(-1)
	oid, users, err := s.userservice.List(stream.Context(), convert.UserContextToDomain(in.UserContext), convert.UserFilterToDomain(in.UserFilter))
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

func (s *UserProtocService) Create(ctx context.Context, in *user.CreateUserRequest) (*user.UserCreatedResponse, error) {
	s.reporter.Increment(1)
	oid, uid, userCID, err := s.userservice.Create(ctx, convert.UserContextToDomain(in.UserContext), convert.UserToDomain(in.User))
	s.reporter.Increment(1)
	if err != nil {
		return nil, err
	}
	return &user.UserCreatedResponse{OrgID: int32(oid), UserID: int32(uid), UserCID: userCID}, nil
}

func (s *UserProtocService) Update(ctx context.Context, in *user.UpdateUserRequest) (*user.UserUpdatedResponse, error) {
	s.reporter.Increment(1)
	oid, uid, err := s.userservice.Update(ctx, convert.UserContextToDomain(in.UserContext), convert.UserToDomain(in.User), int(in.UserID))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &user.UserUpdatedResponse{OrgID: int32(oid), UserID: int32(uid)}, nil
}

func (s *UserProtocService) Delete(ctx context.Context, in *user.DeleteUserRequest) (*user.UserDeletedResponse, error) {
	s.reporter.Increment(1)
	oid, err := s.userservice.Delete(ctx, convert.UserContextToDomain(in.UserContext), int(in.UserID))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &user.UserDeletedResponse{OrgID: int32(oid)}, nil
}

func (s *UserProtocService) AcceptAgreement(ctx context.Context, in *user.UserAgreementRequest) (*user.UserAgreementResponse, error) {
	s.reporter.Increment(1)
	oid, uid, err := s.userservice.AcceptAgreement(ctx, convert.UserContextToDomain(in.UserContext), in.Agreement)
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &user.UserAgreementResponse{OrgID: int32(oid), UserID: int32(uid)}, nil
}
