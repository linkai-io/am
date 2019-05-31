package mock

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
)

type WebDataService struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	AddFn      func(ctx context.Context, userContext am.UserContext, webData *am.WebData) (int, error)
	AddInvoked bool

	GetResponsesFn      func(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.HTTPResponse, error)
	GetResponsesInvoked bool

	GetCertificatesFn      func(ctx context.Context, userContext am.UserContext, filter *am.WebCertificateFilter) (int, []*am.WebCertificate, error)
	GetCertificatesInvoked bool

	GetSnapshotsFn      func(ctx context.Context, userContext am.UserContext, filter *am.WebSnapshotFilter) (int, []*am.WebSnapshot, error)
	GetSnapshotsInvoked bool

	GetURLListFn      func(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.URLListResponse, error)
	GetURLListInvoked bool

	GetDomainDependencyFn      func(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, *am.WebDomainDependency, error)
	GetDomainDependencyInvoked bool

	OrgStatsFn      func(ctx context.Context, userContext am.UserContext) (int, []*am.ScanGroupWebDataStats, error)
	OrgStatsInvoked bool

	GroupStatsFn      func(ctx context.Context, userContext am.UserContext, groupID int) (int, *am.ScanGroupWebDataStats, error)
	GroupStatsInvoked bool

	ArchiveFn      func(ctx context.Context, userContext am.UserContext, group *am.ScanGroup, archiveTime time.Time) (int, int, error)
	ArchiveInvoked bool
}

func (s *WebDataService) Init(config []byte) error {
	return nil
}

func (s *WebDataService) Add(ctx context.Context, userContext am.UserContext, webData *am.WebData) (int, error) {
	s.AddInvoked = true
	return s.AddFn(ctx, userContext, webData)
}

func (s *WebDataService) GetResponses(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.HTTPResponse, error) {
	s.GetResponsesInvoked = true
	return s.GetResponsesFn(ctx, userContext, filter)
}

func (s *WebDataService) GetCertificates(ctx context.Context, userContext am.UserContext, filter *am.WebCertificateFilter) (int, []*am.WebCertificate, error) {
	s.GetCertificatesInvoked = true
	return s.GetCertificatesFn(ctx, userContext, filter)
}

func (s *WebDataService) GetSnapshots(ctx context.Context, userContext am.UserContext, filter *am.WebSnapshotFilter) (int, []*am.WebSnapshot, error) {
	s.GetSnapshotsInvoked = true
	return s.GetSnapshotsFn(ctx, userContext, filter)
}

func (s *WebDataService) GetURLList(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.URLListResponse, error) {
	s.GetURLListInvoked = true
	return s.GetURLListFn(ctx, userContext, filter)
}

func (s *WebDataService) GetDomainDependency(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, *am.WebDomainDependency, error) {
	s.GetDomainDependencyInvoked = true
	return s.GetDomainDependencyFn(ctx, userContext, filter)
}

func (c *WebDataService) OrgStats(ctx context.Context, userContext am.UserContext) (int, []*am.ScanGroupWebDataStats, error) {
	c.OrgStatsInvoked = true
	return c.OrgStatsFn(ctx, userContext)
}

func (c *WebDataService) GroupStats(ctx context.Context, userContext am.UserContext, groupID int) (int, *am.ScanGroupWebDataStats, error) {
	c.GroupStatsInvoked = true
	return c.GroupStatsFn(ctx, userContext, groupID)
}

func (c *WebDataService) Archive(ctx context.Context, userContext am.UserContext, group *am.ScanGroup, archiveTime time.Time) (int, int, error) {
	c.ArchiveInvoked = true
	return c.Archive(ctx, userContext, group, archiveTime)
}
