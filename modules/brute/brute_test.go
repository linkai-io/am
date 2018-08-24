package brute_test

import (
	"os"
	"testing"

	"github.com/linkai-io/am/pkg/dnsclient"

	"github.com/linkai-io/am/modules/brute"
)

func TestAnalyzeZone(t *testing.T) {
	ns := dnsclient.New([]string{"127.0.0.1:2053",
		"8.8.8.8:53",
		"64.6.64.6:53",      // Verisign
		"208.67.222.222:53", // OpenDNS Home
		"77.88.8.8:53",      // Yandex.DNS
		"74.82.42.42:53",    // Hurricane Electric
		"1.0.0.1:53",        // Cloudflare Secondary
		"8.8.4.4:53",        // Google Secondary
		"208.67.220.220:53", // OpenDNS Home Secondary
		"77.88.8.1:53"},     // Yandex.DNS Secondary
		2)

	a := brute.New(ns)
	bruteFile, err := os.Open("testdata/1000.txt")
	if err != nil {
		t.Fatalf("error opening test data file: %s\n", err)
	}

	if err := a.Init(125, bruteFile); err != nil {
		t.Fatalf("error init'ing brute analyzer: %s\n", err)
	}

	//a.AnalyzeZone("linkai.io")
	a.Quit()
}
