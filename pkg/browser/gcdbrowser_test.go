package browser

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

func TestGCDBrowserTLS(t *testing.T) {
	b := NewGCDBrowser(1, 1)
	defer b.Close()
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}
	ctx := context.Background()

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	address := &am.ScanGroupAddress{
		HostAddress: "example.com",
		IPAddress:   "93.184.216.34",
	}

	d, err := b.Load(timeoutCtx, address, "https", "443")
	if err != nil {
		t.Fatalf("error loading: %v\n", err)
	}

	t.Logf("responses: %d\n", len(d.Responses))
	for _, r := range d.Responses {
		t.Logf("RESPONSE: %#v\n", r)
		if r.WebCertificate != nil {
			t.Logf("%#v\n", r.WebCertificate)
		}
	}

	if len(d.Responses) == 0 {
		t.Fatalf("we did not properly intercept and replace host with ip")
	}
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

func TestGCDBrowserTakeScreenshots(t *testing.T) {
	b := NewGCDBrowser(2, 1)
	defer b.Close()
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}
	ctx := context.Background()

	b.SetAPITimeout(time.Second * 30)
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*90)
	defer cancel()

	resultsCh := make(chan *am.WebData, 4)
	defer close(resultsCh)

	go func() {
		address := &am.ScanGroupAddress{
			HostAddress: "google.com",
			IPAddress:   "216.58.197.206",
		}

		w, err := b.Load(timeoutCtx, address, "http", "80")
		if err != nil {
			t.Fatalf("error loading: %v\n", err)
		}
		resultsCh <- w

		address = &am.ScanGroupAddress{
			HostAddress: "example.com",
			IPAddress:   "93.184.216.34",
		}

		w, err = b.Load(timeoutCtx, address, "http", "80")
		if err != nil {
			t.Fatalf("error loading: %v\n", err)
		}
		resultsCh <- w
	}()

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

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	results := 0
	for {
		select {
		case result := <-resultsCh:
			t.Logf("got result: %v\n", result.Snapshot)
			data, _ := base64.StdEncoding.DecodeString(result.Snapshot)
			ioutil.WriteFile(fmt.Sprintf("%d.png", results), data, 0677)
			results++
			if results == 3 {
				return
			}
		case <-timeout.Done():
			t.Fatalf("failed to get results after 20 seconds")
		}
	}
}

func TestGCDBrowserXvfb(t *testing.T) {
	doneContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	xvfbPath := "/usr/bin/Xvfb"
	if _, err := os.Stat(xvfbPath); err != nil {
		t.Logf("not running due to Xvfb not in path")
		return
	}
	go exec.CommandContext(doneContext, xvfbPath, "-ac", ":99", "-screen", "0", "1280x1024x16").Run()
	time.Sleep(500 * time.Microsecond)
	b := NewGCDBrowser(5, 1)
	b.UseDisplay(":99")

	defer b.Close()
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}
	ctx := context.Background()

	b.SetAPITimeout(time.Second * 30)
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*90)
	defer cancel()

	resultsCh := make(chan *am.WebData, 20)
	defer close(resultsCh)

	for i := 0; i < 10; i++ {
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
	}

	go func() {
		address := &am.ScanGroupAddress{
			HostAddress: "google.com",
			IPAddress:   "216.58.197.206",
		}

		w, err := b.Load(timeoutCtx, address, "http", "80")
		if err != nil {
			t.Fatalf("error loading: %v\n", err)
		}
		resultsCh <- w

		address = &am.ScanGroupAddress{
			HostAddress: "example.com",
			IPAddress:   "93.184.216.34",
		}

		w, err = b.Load(timeoutCtx, address, "http", "80")
		if err != nil {
			t.Fatalf("error loading: %v\n", err)
		}
		resultsCh <- w
	}()

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

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	results := 0
	for {
		select {
		case result := <-resultsCh:
			t.Logf("got result: %v\n", result.Snapshot)
			data, _ := base64.StdEncoding.DecodeString(result.Snapshot)
			ioutil.WriteFile(fmt.Sprintf("%d.png", results), data, 0677)
			results++
			if results == 10 {
				return
			}
		case <-timeout.Done():
			t.Fatalf("failed to get results after 20 seconds")
		}
	}
}
