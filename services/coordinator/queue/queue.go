package queue

import (
	"context"

	"github.com/linkai-io/am/am"
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
