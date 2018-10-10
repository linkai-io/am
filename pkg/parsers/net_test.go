package parsers

import (
	"testing"

	"github.com/linkai-io/am/amtest"
)

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
		{"eltd", "amazon.co.uk", 1, false},
		{"eltdherokusub", "app.herokuapp.com", 2, false},
		{"eltdheroku", "herokuapp.com", 1, false},
		{"multi", "test.www.amazon.co.uk", 3, false},
		{"single", "www.amazon.co.uk", 2, false},
		{"multi2", "test1.test2.www.amazon.co.uk", 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDepth(tt.args)
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
			got, err := GetSubDomain(tt.args)
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
		{"single", "www.amazon.co.uk", "www", "amazon.co.uk", false},
		{"multi2", "test1.test2.www.amazon.co.uk", "test1", "test2.www.amazon.co.uk", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSub, gotDomain, err := GetSubDomainAndDomain(tt.args)
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
			got, err := SplitAddresses(tt.args)
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
	etld := SpecialETLD(host)
	if etld != "amazonaws.com" {
		t.Fatalf("error should have returned amazonaws.com for %s\n", host)
	}

	host = "test.com"
	etld = SpecialETLD(host)
	if etld != "test.com" {
		t.Fatalf("error should have returned test.com for %s\n", host)
	}

	host = "test.com."
	etld = SpecialETLD(host)
	if etld != "test.com" {
		t.Fatalf("error should have returned test.com for %s\n", host)
	}
}
