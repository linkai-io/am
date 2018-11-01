package bigdata

import (
	"context"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/dnsclient"
)

func TestBigDataFirstRun(t *testing.T) {
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 2)
	st := amtest.MockBigDataState()
	bds := &mock.BigDataService{}
	bds.GetCTFn = func(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error) {
		return time.Now(), nil, nil
	}
	bds.AddCTFn = func(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error {
		t.Logf("Adding records")
		return nil
	}

	bq := &mock.BigQuerier{}
	bq.QueryETLDFn = func(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error) {
		return amtest.BuildCTRecords(etld, time.Now().UnixNano(), 1), nil
	}
	bd := New(dc, st, bds, bq)
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	address := testBuildAddress("1.1.1.1", "blah.example.com")

	_, newAddrs, err := bd.Analyze(ctx, userContext, address)
	if err != nil {
		t.Fatalf("failed to analyze using big data: %#v\n", err)
	}

	if bq.QueryETLDInvoked == false {
		t.Fatal("query etld should have been invoked")
	}

	if len(newAddrs) != 2 {
		t.Fatalf("failed to find 2 new addresses in big data, got %d\n", len(newAddrs))
	}

}

func TestBigDataRerun(t *testing.T) {
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 2)
	st := amtest.MockBigDataState()
	bds := &mock.BigDataService{}

	// sets lastQueryTime to 1 day ago
	bds.GetCTFn = func(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error) {
		return time.Now().AddDate(0, 0, -1), amtest.BuildCTRecords(etld, time.Now().UnixNano(), 1), nil
	}

	bds.AddCTFn = func(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error {
		t.Logf("Adding records")
		return nil
	}

	bq := &mock.BigQuerier{}
	bq.QueryETLDFn = func(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error) {
		return amtest.BuildCTRecords(etld, time.Now().UnixNano(), 2), nil
	}
	bd := New(dc, st, bds, bq)
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	address := testBuildAddress("1.1.1.1", "blah.example.com")

	_, _, err := bd.Analyze(ctx, userContext, address)
	if err != nil {
		t.Fatalf("failed to analyze using big data: %#v\n", err)
	}

	if bq.QueryETLDInvoked == false {
		t.Fatal("query etld should have been invoked")
	}
}

func TestBigDataNoNewRecords(t *testing.T) {
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 2)
	st := amtest.MockBigDataState()
	bds := &mock.BigDataService{}

	// sets lastQueryTime to 1 day ago
	bds.GetCTFn = func(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error) {
		return time.Now().AddDate(0, 0, -1), amtest.BuildCTRecords(etld, time.Now().UnixNano(), 1), nil
	}

	bds.AddCTFn = func(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error {
		t.Logf("Adding records")
		return nil
	}

	bq := &mock.BigQuerier{}
	bq.QueryETLDFn = func(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error) {
		return amtest.BuildCTRecords(etld, time.Now().UnixNano(), 1), nil
	}
	bd := New(dc, st, bds, bq)
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	address := testBuildAddress("1.1.1.1", "blah.example.com")

	_, _, err := bd.Analyze(ctx, userContext, address)
	if err != nil {
		t.Fatalf("failed to analyze using big data: %#v\n", err)
	}

	if bq.QueryETLDInvoked == false {
		t.Fatal("query etld should have been invoked")
	}

	if bds.AddCTInvoked == true {
		t.Fatal("AddCT should not have been invoked since there are no new records.")
	}
}

func TestBigDataCacheTime(t *testing.T) {
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 2)
	st := amtest.MockBigDataState()
	bds := &mock.BigDataService{}

	// sets lastQueryTime to 1 day ago
	bds.GetCTFn = func(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error) {
		return time.Now(), amtest.BuildCTRecords(etld, time.Now().UnixNano(), 1), nil
	}

	bds.AddCTFn = func(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error {
		t.Logf("Adding records")
		return nil
	}

	bq := &mock.BigQuerier{}
	bq.QueryETLDFn = func(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error) {
		return amtest.BuildCTRecords(etld, time.Now().UnixNano(), 1), nil
	}
	bd := New(dc, st, bds, bq)
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	address := testBuildAddress("1.1.1.1", "blah.example.com")

	_, _, err := bd.Analyze(ctx, userContext, address)
	if err != nil {
		t.Fatalf("failed to analyze using big data: %#v\n", err)
	}

	if bq.QueryETLDInvoked == true {
		t.Fatal("query etld should not have been invoked")
	}

	if bds.AddCTInvoked == true {
		t.Fatal("AddCT should not have been invoked since there are no new records.")
	}
}

func TestBigDataDoCTTime(t *testing.T) {
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 2)
	st := amtest.MockBigDataState()
	bds := &mock.BigDataService{}

	// sets lastQueryTime to 1 day ago
	bds.GetCTFn = func(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error) {
		return time.Now(), amtest.BuildCTRecords(etld, time.Now().UnixNano(), 1), nil
	}

	bds.AddCTFn = func(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error {
		t.Logf("Adding records")
		return nil
	}

	bq := &mock.BigQuerier{}
	bq.QueryETLDFn = func(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error) {
		return amtest.BuildCTRecords(etld, time.Now().UnixNano(), 1), nil
	}
	bd := New(dc, st, bds, bq)
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	address := testBuildAddress("1.1.1.1", "blah.example.com")

	_, _, err := bd.Analyze(ctx, userContext, address)
	if err != nil {
		t.Fatalf("failed to analyze using big data: %#v\n", err)
	}
	_, _, err = bd.Analyze(ctx, userContext, address)
	if err != nil {
		t.Fatalf("failed to analyze using big data: %#v\n", err)
	}
}

func testBuildAddress(ip, host string) *am.ScanGroupAddress {
	addrHash := convert.HashAddress(ip, host)
	return &am.ScanGroupAddress{
		AddressID:           1,
		OrgID:               1,
		GroupID:             1,
		HostAddress:         host,
		IPAddress:           ip,
		DiscoveryTime:       0,
		DiscoveredBy:        "",
		LastScannedTime:     0,
		LastSeenTime:        0,
		ConfidenceScore:     100.0,
		UserConfidenceScore: 0.0,
		IsSOA:               false,
		IsWildcardZone:      false,
		IsHostedService:     false,
		Ignored:             false,
		FoundFrom:           "input_list",
		NSRecord:            0,
		AddressHash:         addrHash,
	}
}
