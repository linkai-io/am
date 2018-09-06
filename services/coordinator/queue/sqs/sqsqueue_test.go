package sqs_test

import (
	"testing"

	"github.com/linkai-io/am/services/coordinator/queue/sqs"
)

func TestCreate(t *testing.T) {
	q := sqs.New("local", "us-faux-1", 5)
	if err := q.Init(); err != nil {
		t.Fatalf("error initializing queue %s\n", err)
	}

	if _, err := q.Create("test"); err != nil {
		t.Fatalf("error creating queue: %s\n", err)
	}
}
