package browser_test

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wirepair/gcd"

	"github.com/linkai-io/am/pkg/browser"
)

func TestGcdLeaser(t *testing.T) {
	l := browser.NewGcdLeaser()

	ts := httptest.NewServer(http.HandlerFunc(l.Acquire))
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", string(port))

	b := gcd.NewChromeDebugger()
	b.ConnectToInstance("localhost", string(port))
	tab, err := b.GetFirstTab()
	if err != nil {
		t.Fatalf("error getting first tab")
	}
	t.Logf("%#v\n", tab.Target)

	tsReturn := httptest.NewServer(http.HandlerFunc(l.Return))
	defer tsReturn.Close()

	res, err = http.Get(tsReturn.URL + "?port=" + string(port))
	if err != nil {
		t.Fatal(err)
	}
	respData, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", string(respData))
	time.Sleep(5 * time.Second)
}

func TestGcdLeaserShutdown(t *testing.T) {
	l := browser.NewGcdLeaser()
	go func() {
		err := l.Serve()
		t.Logf("shutting down %v", err)
	}()
	time.Sleep(1 * time.Second)
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", browser.SOCK)
			},
		},
	}

	resp, err := client.Get("http://unix/acquire")
	if err != nil {
		t.Fatalf("error acquiring browser: %v\n", err)
	}

	port, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if string(port) == "" {
		t.Fatalf("did not get a good port")
	}
	t.Logf("got port %s\n", string(port))

	resp, err = client.Get("http://unix/return?port=" + string(port))
	if err != nil {
		t.Fatalf("error returning browser: %v\n", err)
	}

	respData, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if string(respData) == "" {
		t.Fatalf("did not get a good response")
	}
	t.Logf("got response %s\n", string(respData))
	l.Shutdown()
	time.Sleep(1 * time.Second)
}
