package parsers_test

import (
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/parsers"
)

func TestExtractHostsFromResponse(t *testing.T) {
	needles := make([]*regexp.Regexp, 1)
	needles[0], _ = regexp.Compile("(?i)independent\\.co\\.uk")
	data, err := ioutil.ReadFile("testdata/indep_responses.txt")
	if err != nil {
		t.Fatalf("error opening test file")
	}

	hosts := parsers.ExtractHostsFromResponse(needles, string(data))
	for k, _ := range hosts {
		t.Logf("found: %s\n", k)
	}
	if len(hosts) != 12 {
		t.Fatalf("expected 12 hosts, got: %d\n", len(hosts))
	}

	hosts = parsers.ExtractHostsFromResponse(needles, string("http://www.independent.co.uk"))
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host got: %d\n", len(hosts))
	}

	hosts = parsers.ExtractHostsFromResponse(needles, string("w.independent.co.uk"))
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host got: %d\n", len(hosts))
	}

	hosts = parsers.ExtractHostsFromResponse(needles, string("%w.independent.co.uk"))
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host got: %d\n", len(hosts))
	}
	for k, _ := range hosts {
		t.Logf("%s\n", k)
	}

	hosts = parsers.ExtractHostsFromResponse(needles, string("%windependent.co.uk"))
	if len(hosts) != 0 {
		t.Fatalf("expected 1 host got: %d\n", len(hosts))
	}
}

func TestGetDepth(t *testing.T) {
	type args struct {
		hostAddress string
	}
	tests := []struct {
		name    string
		args    string
		want    int
		wantErr bool
	}{
		{"tld", "co.uk", 0, true},
		{"tld2", "co.id", 0, true},
		{"eltd", "amazon.co.uk", 1, false},
		{"eltdherokusub", "app.herokuapp.com", 2, false},
		{"eltdheroku", "herokuapp.com", 1, false},
		{"multi", "test.www.amazon.co.uk", 3, false},
		{"single", "www.amazon.co.uk", 2, false},
		{"multi2", "test1.test2.www.amazon.co.uk", 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsers.GetDepth(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDepth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != got {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func TestGetSubDomain(t *testing.T) {
	type args struct {
		hostAddress string
	}

	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{"tld", "co.uk", "", true},
		{"eltd", "amazon.co.uk", "", false},
		{"multi", "test.www.amazon.co.uk", "test", false},
		{"single", "www.amazon.co.uk", "www", false},
		{"multi2", "test1.test2.www.amazon.co.uk", "test1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsers.GetSubDomain(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != got {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func TestGetSubDomainAndDomain(t *testing.T) {
	type args struct {
		hostAddress string
	}

	tests := []struct {
		name       string
		args       string
		wantSub    string
		wantDomain string
		wantErr    bool
	}{
		{"tld", "co.uk", "", "", true},
		{"eltd", "amazon.co.uk", "", "amazon.co.uk", false},
		{"multi", "test.www.amazon.co.uk", "test", "www.amazon.co.uk", false},
		{"etld2", "WWW.BTNPROPERTI.CO.ID", "www", "btnproperti.co.id", false},
		{"single", "www.amazon.co.uk", "www", "amazon.co.uk", false},
		{"multi2", "test1.test2.www.amazon.co.uk", "test1", "test2.www.amazon.co.uk", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSub, gotDomain, err := parsers.GetSubDomainAndDomain(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantSub != gotSub {
				t.Errorf("want: %v, got: %v", tt.wantSub, gotSub)
			}

			if tt.wantDomain != gotDomain {
				t.Errorf("want: %v, got: %v", tt.wantDomain, gotDomain)
			}
		})
	}
}

func TestSplitAddresses(t *testing.T) {
	type args struct {
		hostAddress string
	}
	tests := []struct {
		name    string
		args    string
		want    []string
		wantErr bool
	}{
		{"tld", "co.uk", nil, true},
		{"eltd", "amazon.co.uk", nil, false},
		{"multi", "test.www.amazon.co.uk", []string{"www.amazon.co.uk", "amazon.co.uk"}, false},
		{"single", "www.amazon.co.uk", []string{"amazon.co.uk"}, false},
		{"multi2", "test1.test2.www.amazon.co.uk", []string{"amazon.co.uk", "www.amazon.co.uk", "test2.www.amazon.co.uk"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsers.SplitAddresses(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !amtest.SortEqualString(tt.want, got, t) {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func TestSpecialTLD(t *testing.T) {
	host := "test.amazonaws.com"
	etld := parsers.SpecialETLD(host)
	if etld != "amazonaws.com" {
		t.Fatalf("error should have returned amazonaws.com for %s\n", host)
	}

	host = "test.com"
	etld = parsers.SpecialETLD(host)
	if etld != "test.com" {
		t.Fatalf("error should have returned test.com for %s\n", host)
	}

	host = "test.com."
	etld = parsers.SpecialETLD(host)
	if etld != "test.com" {
		t.Fatalf("error should have returned test.com for %s\n", host)
	}

}

func TestBannedIP(t *testing.T) {
	if !parsers.IsBannedIP("127.0.0.1") {
		t.Fatalf("localhost should be banned")
	}

	if parsers.IsBannedIP("129.0.1.1") {
		t.Fatalf("129 should not be banned")
	}

	if parsers.IsBannedIP("2600:9000:20be:de00:16:a8ff:3100:93a1") {
		t.Fatalf("2600:9000:20be:de00:16:a8ff:3100:93a1 should not be banned")
	}
}
