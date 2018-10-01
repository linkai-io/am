package dnsclient

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/linkai-io/am/pkg/parsers"
)

const dnsServer = "127.0.0.53:53"
const localServer = "127.0.0.53:53"

func TestResolveName(t *testing.T) {
	tests := []struct {
		in      string
		isError bool
		outv4   string
		outv6   string
	}{
		{"example.com", false, "93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"},
		{"xn--wgv71a119e.jp", false, "117.104.133.167", "2001:218:3001:7::80"},
		{"日本語.jp", false, "117.104.133.167", "2001:218:3001:7::80"},
		{"thisisnotarealdomainthisisatestsomethingsomething123.com", true, "", ""},
		{"1.1", true, "", ""},
	}

	c := New([]string{dnsServer}, 3)
	for _, test := range tests {
		ctx := context.Background()
		r, err := c.ResolveName(ctx, test.in)
		if err != nil {
			if !test.isError {
				t.Fatalf("%s error: %s\n", test.in, err)
			}
			continue
		}

		for _, rr := range r {
			if test.isError {
				t.Fatalf("%s expected error, did not get one.\n", test.in)
			}
			for _, ip := range rr.IPs {
				p := net.ParseIP(ip)
				if p == nil {
					t.Fatalf("did not get a valid IP address back, got: %s\n", ip)
				}
				if strings.Compare(ip, test.outv6) != 0 && strings.Compare(ip, test.outv4) != 0 {
					t.Fatalf("expected %s or %s, got %s\n", test.outv6, test.outv4, ip)
				}
			}
			t.Logf("%#v\n", r)
		}

	}
}

func TestResolveIPv4(t *testing.T) {
	tests := []struct {
		in      string
		isError bool
		out     string
	}{
		{"1.0.0.1", false, "one.one.one.one"},
		{"1.1.1.1", false, "one.one.one.one"},
		{"192.168.0.1", true, ""},
		{"1.1", true, ""},
	}
	c := New([]string{dnsServer}, 3)
	for _, test := range tests {
		ctx := context.Background()
		r, err := c.ResolveIP(ctx, test.in)

		if err != nil {
			if !test.isError {
				t.Fatalf("%s error: %s\n", test.in, err)
			}
			continue
		}

		if test.isError {
			t.Fatalf("%s expected error, did not get one.\n", test.in)
		}

		if r.Hosts[0] != test.out {
			t.Fatalf("%s expected %s got %s\n", test.in, test.out, r.Hosts[0])
		}
		t.Logf("%#v\n", r)
	}
}

func TestResolveIPv6(t *testing.T) {
	tests := []struct {
		in      string
		isError bool
		out     string
	}{
		{"2606:4700:4700::1001", false, "one.one.one.one"},
		{"2606:4700:4700::1111", false, "one.one.one.one"},
		{"2404:6800:4004:80d::2004", false, "nrt12s17-in-x04.1e100.net"},
		{"2606:4700:", true, ""},
		{"dead:beef::", true, ""},
	}

	c := New([]string{dnsServer}, 3)
	for _, test := range tests {
		ctx := context.Background()
		r, err := c.ResolveIP(ctx, test.in)

		if err != nil {
			if !test.isError {
				t.Fatalf("%s error: %s\n", test.in, err)
			}
			continue
		}

		if test.isError {
			t.Fatalf("%s expected error, did not get one.\n", test.in)
		}

		if r.Hosts[0] != test.out {
			t.Fatalf("%s expected %s got %s\n", test.in, test.out, r.Hosts[0])
		}
		t.Logf("%#v\n", r)
	}
}

func TestLookupNS(t *testing.T) {
	tests := []struct {
		in      string
		isError bool
		outlen  int
	}{
		{"zonetransfer.me", false, 2},
		{"linkai.io", false, 4},
		{"google.com", false, 4},
		{"invalid.linkai.io", true, 0},
		{"0932jzzzzzzzzzzzzz.com", true, 0},
	}

	c := New([]string{dnsServer}, 3)

	for _, test := range tests {
		ctx := context.Background()
		r, err := c.LookupNS(ctx, test.in)

		if err != nil {
			if !test.isError {
				t.Fatalf("%s error: %s\n", test.in, err)
			}
			continue
		}

		if test.isError {
			t.Fatalf("%s expected error, did not get one.\n", test.in)
		}

		if len(r.Hosts) != test.outlen {
			t.Fatalf("%s expected %d records got %d\n", test.in, test.outlen, len(r.Hosts))
		}
		t.Logf("%#v\n", r)
	}

}

func TestDoAXFR(t *testing.T) {
	c := New([]string{dnsServer}, 3)
	ctx := context.Background()
	r, err := c.DoAXFR(ctx, "zonetransfer.me")
	if err != nil {
		t.Fatalf("error: %s\n", err)
	}
	for ns, axfr := range r {
		for _, results := range axfr {
			t.Logf("ns: %s %#v %d %d %s\n", ns, results, len(results.Hosts), len(results.IPs), results.Type())
		}
	}
	ctx = context.Background()
	r, err = c.DoAXFR(ctx, "linkai.io")
	if err != nil {
		t.Fatalf("error %s\n", err)
	}

	for ns, axfr := range r {
		t.Logf("%s %#v\n", ns, axfr)
	}

}

func TestParseArpa(t *testing.T) {
	tests := []struct {
		in   string
		isOK bool
		out  string
	}{
		{"5.2.0.192.in-addr.arpa.", true, "192.0.2.5"},
		{"4.3.3.7.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.1.1.1.1.0.0.2.ip6.arpa", true, "2001:1111:0000:0000:0000:0000:0000:7334"},
		{"4.3.3.7.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.1.1.1.1.a.0.2.ip6.arpa", true, "20a1:1111:0000:0000:0000:0000:0000:7334"},
		{"4.3.3.7.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.1.1.1.1.0.2.ip6.arpa", false, ""},
		{"5.2.0.in-addr.arpa.", false, ""},
		{"a.2.0.192.in-addr.arpa.", false, ""},
		{"5.2.0.192in-addr.arpa.", false, ""},
	}
	for _, test := range tests {
		ip, ok := parsers.ParseArpa(test.in)
		if test.isOK && test.out != ip {
			t.Fatalf("error expected %s got %s\n", test.out, ip)
		}

		if ok && !test.isOK {
			t.Fatalf("error got ok when test was not supposed to return ok for %s output: %s\n", test.in, ip)
		}
	}
}
