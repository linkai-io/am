package mock

import (
	"context"

	"github.com/linkai-io/am/am"
)

type Storage struct {
	InitFn      func() error
	InitInvoked bool

	WriteFn      func(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, data []byte) (string, string, error)
	WriteInvoked bool
}

func (s *Storage) Init() error {
	s.InitInvoked = true
	return s.InitFn()
}

func (s *Storage) Write(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, data []byte) (string, string, error) {
	s.WriteInvoked = true
	return s.WriteFn(ctx, userContext, address, data)
}
