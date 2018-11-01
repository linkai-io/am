package mock

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
)

type BigQuerier struct {
	InitFn func(config []byte) error

	QueryETLDFn      func(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error)
	QueryETLDInvoked bool
}

func (b *BigQuerier) Init(config []byte) error {
	return nil
}

func (b *BigQuerier) QueryETLD(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error) {
	b.QueryETLDInvoked = true
	return b.QueryETLDFn(ctx, from, etld)
}
