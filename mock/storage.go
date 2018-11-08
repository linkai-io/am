package mock

import (
	"context"

	"github.com/linkai-io/am/pkg/secrets"

	"github.com/linkai-io/am/am"
)

type Storage struct {
	InitFn      func(cache *secrets.SecretsCache) error
	InitInvoked bool

	WriteFn      func(ctx context.Context, address *am.ScanGroupAddress, data []byte) (string, string, error)
	WriteInvoked bool
}

func (s *Storage) Init(cache *secrets.SecretsCache) error {
	s.InitInvoked = true
	return s.InitFn(cache)
}

func (s *Storage) Write(ctx context.Context, address *am.ScanGroupAddress, data []byte) (string, string, error) {
	s.WriteInvoked = true
	return s.WriteFn(ctx, address, data)
}
