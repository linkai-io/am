package mock

import (
	"context"

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
