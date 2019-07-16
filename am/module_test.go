package am_test

import (
	"testing"

	"github.com/linkai-io/am/amtest"
)

func TestPortModule(t *testing.T) {
	cfg := amtest.CreateModuleConfig()
	/*
		allowedTLDs := []string{"example.com"}
		disallowedTLDs := []string{"blah.com"}
		allowedHosts := []string{"scanme.blah.com"}
		disallowedHosts := []string{"noportscan.example.com"}
	*/
	tests := []struct {
		etld string
		host string
		pass bool
	}{
		{etld: "blah.com", host: "scanme.blah.com", pass: true},
		{etld: "blah.com", host: "blah.com", pass: false},
		{etld: "blah.com", host: "www.blah.com", pass: false},
		{etld: "example.com", host: "example.com", pass: true},
		{etld: "example.com", host: "www.example.com", pass: true},
		{etld: "example.com", host: "noportscan.example.com", pass: false},
	}

	for _, tt := range tests {
		got := cfg.PortModule.CanPortScan(tt.etld, tt.host)
		if got != tt.pass {
			t.Fatalf("etld: %v host: %v expected %v got %v\n", tt.etld, tt.host, tt.pass, got)
		}
	}

	ipTests := []struct {
		ip   string
		pass bool
	}{
		{ip: "192.168.1.1", pass: true},
		{ip: "", pass: false},
		{ip: "10.1.1.1", pass: false},
	}
	cfg.PortModule.DisallowedHosts = append(cfg.PortModule.DisallowedHosts, "10.1.1.1")

	for _, tt := range ipTests {
		got := cfg.PortModule.CanPortScanIP(tt.ip)
		if got != tt.pass {
			t.Fatalf("ip: %v expected %v got %v\n", tt.ip, tt.pass, got)
		}
	}
}
