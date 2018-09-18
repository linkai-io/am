package retrier_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/linkai-io/am/pkg/retrier"
)

func TestRetryUntil(t *testing.T) {
	expected := 6
	attempts := 0

	x := func() error {
		fmt.Printf("tick %d\n", attempts)
		attempts++
		return errors.New("")
	}
	_ = retrier.RetryUntil(x, time.Second*5, time.Second*1)
	if attempts != expected {
		t.Fatalf("expected %d got %d\n", expected, attempts)
	}
}
