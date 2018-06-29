package scangroup

import (
	"context"

	"gopkg.linkai.io/v1/repos/am/services/scangroup/protoc"
	"gopkg.linkai.io/v1/repos/am/services/scangroup/store"
)

type Service struct {
	store store.Storer
}

func New(store store.Storer) *Service {
	return &Service{store: store}
}

func (s *Service) Init(config []byte) error {

	return nil
}

func (s *Service) Create(ctx context.Context, in *protoc.NewGroupRequest) (*protoc.VersionCreatedResponse, error) {
	return &protoc.VersionCreatedResponse{GroupID: 1, GroupVersionID: 1}, nil
}

func (s *Service) Delete(ctx context.Context, in *protoc.DeleteGroupRequest) (*protoc.GroupDeletedResponse, error) {
	return &protoc.GroupDeletedResponse{OrgID: 1, GroupID: 1}, nil
}

func (s *Service) GetVersion(ctx context.Context, in *protoc.VersionRequest) (*protoc.GroupVersion, error) {
	return &protoc.GroupVersion{OrgID: 1, GroupID: 1, GroupVersionID: 1}, nil
}

func (s *Service) CreateVersion(ctx context.Context, in *protoc.NewVersionRequest) (*protoc.VersionCreatedResponse, error) {
	return &protoc.VersionCreatedResponse{GroupID: 1, GroupVersionID: 1}, nil
}

func (s *Service) DeleteVersion(ctx context.Context, in *protoc.DeleteVersionRequest) (*protoc.VersionDeletedResponse, error) {
	return &protoc.VersionDeletedResponse{GroupID: 1, GroupVersionID: 1}, nil
}

func (s *Service) Groups(in *protoc.GroupsRequest, stream protoc.ScanGroup_GroupsServer) error {
	return nil
}

func (s *Service) Get(ctx context.Context, in *protoc.GroupRequest) (*protoc.Group, error) {
	return &protoc.Group{}, nil
}
