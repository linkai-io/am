package parsers

import (
	"testing"

	"github.com/linkai-io/am/amtest"
)

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
