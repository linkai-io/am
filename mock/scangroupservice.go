package mock

import (
	"context"

	"github.com/linkai-io/am/am"
)

type ScanGroupService struct {
	InitFn      func(config []byte) error
	InitInvoked bool

	GetFn      func(ctx context.Context, userContext am.UserContext, groupID int) (oid int, group *am.ScanGroup, err error)
	GetInvoked bool

	GetByNameFn      func(ctx context.Context, userContext am.UserContext, groupName string) (oid int, group *am.ScanGroup, err error)
	GetByNameInvoked bool

	AllGroupsFn      func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupFilter) (groups []*am.ScanGroup, err error)
	AllGroupsInvoked bool

	GroupsFn      func(ctx context.Context, userContext am.UserContext) (oid int, groups []*am.ScanGroup, err error)
	GroupsInvoked bool

	CreateFn      func(ctx context.Context, userContext am.UserContext, newGroup *am.ScanGroup) (oid int, gid int, err error)
	CreateInvoked bool

	CountFn      func(ctx context.Context, userContext am.UserContext, groupID int) (oid int, count int, err error)
	CountInvoked bool

	UpdateFn      func(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) (oid int, gid int, err error)
	UpdateInvoked bool

	DeleteFn      func(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error)
	DeleteInvoked bool

	PauseFn      func(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error)
	PauseInvoked bool

	ResumeFn      func(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error)
	ResumeInvoked bool
}

func (s *ScanGroupService) Init(config []byte) error {
	s.InitInvoked = true
	return s.InitFn(config)
}

func (s *ScanGroupService) Get(ctx context.Context, userContext am.UserContext, groupID int) (oid int, group *am.ScanGroup, err error) {
	s.GetInvoked = true
	return s.GetFn(ctx, userContext, groupID)
}

func (s *ScanGroupService) GetByName(ctx context.Context, userContext am.UserContext, groupName string) (oid int, group *am.ScanGroup, err error) {
	s.GetByNameInvoked = true
	return s.GetByNameFn(ctx, userContext, groupName)
}

func (s *ScanGroupService) AllGroups(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupFilter) (groups []*am.ScanGroup, err error) {
	s.AllGroupsInvoked = true
	return s.AllGroupsFn(ctx, userContext, filter)
}

func (s *ScanGroupService) Groups(ctx context.Context, userContext am.UserContext) (oid int, groups []*am.ScanGroup, err error) {
	s.GroupsInvoked = true
	return s.GroupsFn(ctx, userContext)
}

func (s *ScanGroupService) Create(ctx context.Context, userContext am.UserContext, newGroup *am.ScanGroup) (oid int, gid int, err error) {
	s.CreateInvoked = true
	return s.CreateFn(ctx, userContext, newGroup)
}

func (s *ScanGroupService) Update(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) (oid int, gid int, err error) {
	s.UpdateInvoked = true
	return s.UpdateFn(ctx, userContext, group)
}

func (s *ScanGroupService) Delete(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	s.DeleteInvoked = true
	return s.DeleteFn(ctx, userContext, groupID)
}

func (s *ScanGroupService) Pause(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	s.PauseInvoked = true
	return s.PauseFn(ctx, userContext, groupID)
}

func (s *ScanGroupService) Resume(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	s.ResumeInvoked = true
	return s.ResumeFn(ctx, userContext, groupID)
}
