package browser

import (
	"context"
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
)

func TestGCDBrowserPool(t *testing.T) {
	ctx := context.Background()
	b := NewGCDBrowserPool(5, amtest.MockWebDetector())
	defer b.Close(ctx)

	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}

	address := &am.ScanGroupAddress{
		HostAddress: "example.com",
		IPAddress:   "93.184.216.34",
	}

	webData, err := b.Load(ctx, address, "http", "80")
	if err != nil {
		t.Fatalf("error during load: %v\n", err)
	}
	for _, resp := range webData.Responses {
		t.Logf("%v\n", resp.URL)
	}
}

func TestGCDBrowserPoolTLS(t *testing.T) {
	ctx := context.Background()
	b := NewGCDBrowserPool(1, amtest.MockWebDetector())
	defer b.Close(ctx)
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	address := &am.ScanGroupAddress{
		//HostAddress: "example.com",
		HostAddress: "example.com",
		IPAddress:   "93.184.216.34",
		//IPAddress: "192.168.62.130",
		//HostAddress: "www.veracode.com",
		//IPAddress:   "104.17.7.6",
	}

	d, err := b.Load(timeoutCtx, address, "https", "443")
	if err != nil {
		t.Fatalf("error loading: %v\n", err)
	}

	t.Logf("responses: %d\n", len(d.Responses))
	for _, r := range d.Responses {
		//t.Logf("RESPONSE: %#v\n", r)
		if r.WebCertificate != nil {
			t.Logf("%#v\n", r.WebCertificate)
		}
	}

	if len(d.Responses) == 0 {
		t.Fatalf("we did not properly intercept and replace host with ip")
	}
}

func TestGCDBrowserPoolClosedPort(t *testing.T) {
	ctx := context.Background()
	b := NewGCDBrowserPool(1, amtest.MockWebDetector())
	defer b.Close(ctx)
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	address := &am.ScanGroupAddress{
		HostAddress: "example.com",
		IPAddress:   "93.184.216.34",
	}

	_, err := b.Load(timeoutCtx, address, "http", "8555")
	if err == nil {
		t.Fatalf("did not get expected error")
	}
}

func TestGCDBrowserPoolNavFailure(t *testing.T) {

	ctx := context.Background()
	b := NewGCDBrowserPool(2, amtest.MockWebDetector())
	b.SetAPITimeout(time.Second * 3)
	defer b.Close(ctx)

	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}

	address := &am.ScanGroupAddress{
		HostAddress: "fjffjfjfjfjfjfjfjfjfjfow9ioiwrowir.com",
		IPAddress:   "240.0.0.1",
	}

	_, err := b.Load(ctx, address, "http", "80")
	if err == nil {
		t.Fatalf("did not get error when it was expected\n")
	}
}

func TestGCDBrowserPoolTakeScreenshots(t *testing.T) {
	ctx := context.Background()
	b := NewGCDBrowserPool(2, amtest.MockWebDetector())
	defer b.Close(ctx)

	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}

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
		case <-resultsCh:
			//t.Logf("got result: %v\n", result.Snapshot)
			//data, _ := base64.StdEncoding.DecodeString(result.Snapshot)
			//ioutil.WriteFile(fmt.Sprintf("%d.png", results), data, 0677)
			results++
			if results == 3 {
				return
			}
		case <-timeout.Done():
			t.Fatalf("failed to get results after 20 seconds")
		}
	}
}
