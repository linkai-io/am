package retrier_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/linkai-io/am/am"

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

func TestRetryUnless(t *testing.T) {
	errType := am.ErrUserNotAuthorized
	expected := 2
	attempts := 0
	x := func() error {
		attempts++
		if attempts == 1 {
			return am.ErrEmptyAddress
		}
		return am.ErrUserNotAuthorized
	}
	err := retrier.RetryUnless(x, errType)
	if err == nil {
		t.Fatalf("fail")
	}
	if expected != attempts {
		t.Fatalf("expected %v attempts got %v\n", expected, attempts)
	}
	t.Logf("error: %v\n", err)
}
