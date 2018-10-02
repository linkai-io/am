package brute_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/dnsclient"

	"github.com/linkai-io/am/services/module/brute"
)

func TestAnalyze(t *testing.T) {
	orgID := 1
	userID := 1
	groupID := 1
	input, err := os.Open("testdata/10.txt")
	if err != nil {
		t.Fatalf("error opening input: %s\n", err)
	}
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 1)
	st := amtest.MockBruteState()

	b := brute.New(dc, st)
	if err := b.Init(input); err != nil {
		t.Fatalf("error initializing brute forcer: %v\n", err)
	}

	addrs := amtest.AddrsFromInputFile(orgID, groupID, strings.NewReader("linkai.io"), t)
	// special case because herokuapp.com matches a TLD
	heroku := amtest.AddrsFromInputFile(orgID, groupID, strings.NewReader("ignore.herokuapp.com"), t)
	heroku[0].HostAddress = "herokuapp.com"
	heroku[0].AddressHash = convert.HashAddress("", "herokuapp.com")

	zonetransferme := amtest.AddrsFromInputFile(orgID, groupID, strings.NewReader("zonetransfer.me"), t)
	mutate := amtest.AddrsFromInputFile(orgID, groupID, strings.NewReader("intns1.zonetransfer.me"), t)

	ctx := context.Background()
	userContext := amtest.CreateUserContext(orgID, userID)

	tests := []struct {
		in           *am.ScanGroupAddress
		isError      bool
		isWildcard   bool
		hasResultLen int
	}{
		{addrs[0], false, false, 12},
		{addrs[0], false, false, 0}, // second check should be ignored and return 0 records because it's in cache
		{heroku[0], false, true, 0},
		{zonetransferme[0], false, false, 1},
		{mutate[0], false, false, 1},
	}

	for _, test := range tests {
		original, results, err := b.Analyze(ctx, userContext, test.in)
		if err != nil {
			if !test.isError {
				t.Fatalf("%v error: %v\n", test.in, err)
			}
			continue
		}

		if original.IsWildcardZone != test.isWildcard {
			t.Fatalf("expected wildcard %v got %v\n", test.isWildcard, original.IsWildcardZone)
		}

		if len(results) != test.hasResultLen {
			t.Fatalf("%v expected %d results got %d\n", test.in.HostAddress, test.hasResultLen, len(results))
		}
	}

}
func TestBuildSubDomainList(t *testing.T) {
	expected := 4
	list := []string{"a", "b"}
	custom := []string{"c", "d"}
	results := brute.BuildSubDomainList(list, custom)
	if len(results) != expected {
		t.Fatalf("did not get proper size back, expected %d got %d\n", expected, len(results))
	}
	expectedDomains := []string{"a", "b", "c", "d"}
	if !amtest.SortEqualString(expectedDomains, results, t) {
		t.Fatalf("expected %v got %v\n", expectedDomains, results)
	}

	// test with empty custom domains
	results = brute.BuildSubDomainList(expectedDomains, []string{})
	if len(results) != expected {
		t.Fatalf("did not get proper size back, expected %d got %d\n", expected, len(results))
	}
}
