package inputlist

import (
	"strings"
	"testing"
)

const testMaxAddresses = 500

func TestParseListCIDR(t *testing.T) {
	numErrors := 3
	numAddrs := 264
	lines := `192.168.4
	12:3456:78:90ab:cd:ef01:23:30/125
	192.168.2.0/24
	2001:db8:a0b:12f0::1/32
	1.0.0.0/2`

	r := strings.NewReader(lines)
	addr, errs := ParseList(r, testMaxAddresses)
	t.Logf("errors: %d\n", len(errs))
	if len(errs) != numErrors {
		for _, err := range errs {
			t.Logf("%#v %s\n", err, err.Err)
		}
		t.Fatalf("expected %d errors got: %d", numErrors, len(errs))
	}

	if len(addr) != numAddrs {
		t.Fatalf("expected %d addresses got: %d\n", numAddrs, len(addr))
	}

	for _, err := range errs {
		t.Logf("%#v %s\n", err, err.Err)
	}

	t.Logf("addrs: %d\n", len(addr))
}

func TestParseListURL(t *testing.T) {
	numErrors := 2
	numAddrs := 8
	lines := `https://日本語.com
	https://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]
	http://google.com/
	ftp://ftp.example.com
	http://co.uk/
	http://10.1.1.1/
	http://*.domain.com
	https://www.example.com:9090/asdf/?asdf=qwer#foo=bar;baz=boo
	http://example.com.
	http:///hi`

	r := strings.NewReader(lines)
	addr, errs := ParseList(r, testMaxAddresses)

	if len(errs) != numErrors {
		t.Fatalf("expected %d errors got: %d", numErrors, len(errs))
	}

	if len(addr) != numAddrs {
		t.Fatalf("expected %d addresses got: %d\n", numAddrs, len(addr))
	}

	for host := range addr {
		t.Logf("%s\n", host)
	}
}

func TestParseListIPHost(t *testing.T) {
	numErrors := 5
	numHosts := 7
	lines := `192.168.2.1
	2001:0db8:85a3:0000:0000:8a2e:0370:7334
	2001:0db8:85a3:0000:0000:8a2e:0370
	asdf
	.com
	co.uk
	asdf.com
	日本語.com
	..example1.com
	*.domain.com
	example.com.
	["test.linkai.io"],["blah.linkai.io"]`

	r := strings.NewReader(lines)
	addr, errs := ParseList(r, testMaxAddresses)
	t.Logf("errors: %d\n", len(errs))

	for _, err := range errs {
		t.Logf("%#v %s\n", err, err.Err)
	}

	if len(errs) != numErrors {
		t.Fatalf("expected %d errors got: %d", numErrors, len(errs))
	}

	if len(addr) != numHosts {
		t.Fatalf("expected %d hosts got: %d\n", numHosts, len(addr))
	}

	for host := range addr {
		t.Logf("%s\n", host)
	}
}

func TestParseListMaxAddresses(t *testing.T) {

	lines := `192.168.2.1
	asdf
	asdf.com
	日本語.com
	..example1.com
	*.domain.com
	example.com.`

	r := strings.NewReader(lines)
	_, errs := ParseList(r, 4)
	if errs[len(errs)-1].Err != ErrTooManyAddresses.Error() {
		t.Fatalf("expected last err to be too many addresses")
	}

	lines = `192.168.1.0/24
	192.168.2.0/24`

	r = strings.NewReader(lines)
	_, errs = ParseList(r, 256)
	if errs[len(errs)-1].Err != ErrTooManyAddresses.Error() {
		t.Fatalf("expected err due to max address with cidr")
	}

	if errs[0].LineNumber != 2 {
		t.Fatalf("expected error to be encountered on second line")
	}

	lines = `http://example.com
	http://blah.com`

	r = strings.NewReader(lines)
	_, errs = ParseList(r, 1)
	if errs[len(errs)-1].Err != ErrTooManyAddresses.Error() {
		t.Fatalf("expected err due to max address with cidr")
	}

	if errs[0].LineNumber != 2 {
		t.Fatalf("expected error to be encountered on second line")
	}
}
