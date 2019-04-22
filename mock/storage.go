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

	WriteWithHashFn      func(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, data []byte, hashName string) (string, error)
	WriteWithHashInvoked bool

	GetInfraFileFn      func(ctx context.Context, bucketName, objectName string) ([]byte, error)
	GetFileInfraInvoked bool

	PutInfraFileFn      func(ctx context.Context, bucketName, objectName string, data []byte) error
	PutInfraFileInvoked bool
}

func (s *Storage) Init() error {
	s.InitInvoked = true
	return s.InitFn()
}

func (s *Storage) Write(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, data []byte) (string, string, error) {
	s.WriteInvoked = true
	return s.WriteFn(ctx, userContext, address, data)
}

func (s *Storage) WriteWithHash(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, data []byte, hashName string) (string, error) {
	s.WriteWithHashInvoked = true
	return s.WriteWithHashFn(ctx, userContext, address, data, hashName)
}

func (s *Storage) GetInfraFile(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	s.GetFileInfraInvoked = true
	return s.GetInfraFileFn(ctx, bucketName, objectName)
}

func (s *Storage) PutInfraFile(ctx context.Context, bucketName, objectName string, data []byte) error {
	s.PutInfraFileInvoked = true
	return s.PutInfraFileFn(ctx, bucketName, objectName, data)
}
