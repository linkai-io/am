package browser

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/linkai-io/am/am"
)

func TestGCDBrowser(t *testing.T) {
	b := NewGCDBrowser(5, 3)
	defer b.Close()
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}

	address := &am.ScanGroupAddress{
		HostAddress: "independent.co.uk",
		IPAddress:   "151.101.65.184",
	}

	webData, err := b.Load(context.Background(), address, "http", "80")
	if err != nil {
		t.Fatalf("error during load: %v\n", err)
	}
	for _, resp := range webData.Responses {
		t.Logf("%v\n", resp.URL)
	}
	t.Logf("sleeping...\n")
	time.Sleep(time.Second * 1)
}

func TestGCDBrowserFailure(t *testing.T) {
	b := NewGCDBrowser(2, 1)
	defer b.Close()
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}
	ctx := context.Background()

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	address := &am.ScanGroupAddress{
		HostAddress: "example.com",
		IPAddress:   "93.184.216.34",
	}
	// fake crashed tabs
	atomic.AddInt32(&b.tabErrors, 2)

	if _, err := b.Load(timeoutCtx, address, "http", "80"); err != nil {
		t.Fatalf("error loading: %v\n", err)
	}

}

func TestGCDBrowserNavFailure(t *testing.T) {
	b := NewGCDBrowser(2, 1)
	b.SetAPITimeout(time.Second * 3)
	defer b.Close()
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}
	ctx := context.Background()

	address := &am.ScanGroupAddress{
		HostAddress: "fjffjfjfjfjfjfjfjfjfjfow9ioiwrowir.com",
		IPAddress:   "240.0.0.1",
	}

	_, err := b.Load(ctx, address, "http", "80")
	if err == nil {
		t.Fatalf("did not get error when it was expected\n")
	}
}

// TestGCDBrowserFailureHandlingMultiTabs tests that a browser in the middle
// of shutting down will still complete its tasks.
func TestGCDBrowserFailureHandlingMultiTabs(t *testing.T) {
	b := NewGCDBrowser(2, 1)
	defer b.Close()
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}
	ctx := context.Background()

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*90)
	defer cancel()

	resultsCh := make(chan *am.WebData, 2)

	go func() {
		address := &am.ScanGroupAddress{
			HostAddress: "independent.co.uk",
			IPAddress:   "151.101.65.184",
		}

		w, err := b.Load(timeoutCtx, address, "http", "80")
		if err != nil {
			t.Fatalf("error loading: %v\n", err)
		}
		resultsCh <- w
	}()
	// fake crashed tabs
	time.Sleep(200 * time.Millisecond)
	atomic.AddInt32(&b.tabErrors, 2)

	// run second which should cause restart signal
	go func() {
		address := &am.ScanGroupAddress{
			HostAddress: "example.com",
			IPAddress:   "93.184.216.34",
		}

		w, err := b.Load(timeoutCtx, address, "http", "80")
		if err != nil {
			t.Fatalf("error loading: %v\n", err)
		}
		resultsCh <- w
	}()

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	results := 0
	for {
		select {
		case result := <-resultsCh:
			t.Logf("got result: %#v\n", result)
			results++
			if results == 2 {
				return
			}
		case <-timeout.Done():
			t.Fatalf("failed to get results after 20 seconds")
		}
	}
}
