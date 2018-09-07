package queue

import (
	"context"

	"github.com/linkai-io/am/am"
)

// Queue interface for managing queues
type Queue interface {
	List() (map[string]string, error)
	Create(name string) (string, error)
	Pause(queue string) error
	Delete(queue string) error
	Stats(queue string) error
	PushAddresses(ctx context.Context, queue string, addresses []*am.ScanGroupAddress) error
}
