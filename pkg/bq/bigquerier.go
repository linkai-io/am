package bq

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
)

type BigQuerier interface {
	QueryETLD(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error)
}
