package mock

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
)

type BigQuerier struct {
	InitFn func(config, credentials []byte) error

	QueryETLDFn      func(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error)
	QueryETLDInvoked bool

	QuerySubdomainsFn      func(ctx context.Context, from time.Time, etld string) (map[string]*am.CTSubdomain, error)
	QuerySubdomainsInvoked bool
}

func (b *BigQuerier) Init(config, credentials []byte) error {
	return nil
}

func (b *BigQuerier) QueryETLD(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error) {
	b.QueryETLDInvoked = true
	return b.QueryETLDFn(ctx, from, etld)
}

func (b *BigQuerier) QuerySubdomains(ctx context.Context, from time.Time, etld string) (map[string]*am.CTSubdomain, error) {
	b.QuerySubdomainsInvoked = true
	return b.QuerySubdomainsFn(ctx, from, etld)
}
