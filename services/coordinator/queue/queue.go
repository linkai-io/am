package queue

import (
	"context"

	"gopkg.linkai.io/v1/repos/am/am"
)

// Queue interface for managing queues
type Queue interface {
	List() (map[string]string, error)
	Create(name string) error
	Pause(name string) error
	Delete(name string) error
	Stats(name string) error
	PushAddresses(ctx context.Context, addresses []*am.ScanGroupAddress) error
}
