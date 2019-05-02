package mock

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
)

type BigDataService struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	GetCTFn      func(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error)
	GetCTInvoked bool

	AddCTFn      func(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error
	AddCTInvoked bool

	DeleteCTFn      func(ctx context.Context, userContext am.UserContext, etld string) error
	DeleteCTInvoked bool

	GetCTSubdomainsFn      func(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTSubdomain, error)
	GetCTSubdomainsInvoked bool

	AddCTSubdomainsFn      func(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, subdomains map[string]*am.CTSubdomain) error
	AddCTSubdomainsInvoked bool

	DeleteCTSubdomainsFn      func(ctx context.Context, userContext am.UserContext, etld string) error
	DeleteCTSubdomainsInvoked bool

	GetETLDsFn      func(ctx context.Context, userContext am.UserContext) ([]*am.CTETLD, error)
	GetETLDSInvoked bool
}

func (s *BigDataService) Init(config []byte) error {
	return nil
}

func (s *BigDataService) GetETLDs(ctx context.Context, userContext am.UserContext) ([]*am.CTETLD, error) {
	s.GetETLDSInvoked = true
	return s.GetETLDsFn(ctx, userContext)
}

func (s *BigDataService) GetCT(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error) {
	s.GetCTInvoked = true
	return s.GetCTFn(ctx, userContext, etld)
}

func (s *BigDataService) AddCT(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error {
	s.AddCTInvoked = true
	return s.AddCTFn(ctx, userContext, etld, queryTime, ctRecords)
}

func (s *BigDataService) DeleteCT(ctx context.Context, userContext am.UserContext, etld string) error {
	s.DeleteCTInvoked = true
	return s.DeleteCTFn(ctx, userContext, etld)
}

func (s *BigDataService) GetCTSubdomains(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTSubdomain, error) {
	s.GetCTSubdomainsInvoked = true
	return s.GetCTSubdomainsFn(ctx, userContext, etld)
}

func (s *BigDataService) AddCTSubdomains(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, subdomains map[string]*am.CTSubdomain) error {
	s.AddCTSubdomainsInvoked = true
	return s.AddCTSubdomainsFn(ctx, userContext, etld, queryTime, subdomains)
}

func (s *BigDataService) DeleteCTSubdomains(ctx context.Context, userContext am.UserContext, etld string) error {
	s.DeleteCTSubdomainsInvoked = true
	return s.DeleteCTSubdomainsFn(ctx, userContext, etld)
}
