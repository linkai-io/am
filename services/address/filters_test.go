package address

import (
	"testing"
	"time"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/amtest"
)

func TestBuildGetFilterQuery(t *testing.T) {
	userContext := amtest.CreateUserContext(1, 1)
	filter := &am.ScanGroupAddressFilter{
		OrgID:   userContext.GetOrgID(),
		GroupID: 1,
		Start:   0,
		Limit:   1000,
		Filters: &am.FilterType{},
	}
	query, args, err := buildGetFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building get query: %v\n", err)
	}
	if len(args) != 4 {
		t.Fatalf("invalid number of args expected 4 got: %d\n", len(args))
	}
	expected := "SELECT sga.organization_id,  sga.address_id,  sga.scan_group_id,  sga.host_address, sga.ip_address,  sga.discovered_timestamp,  (select discovered_by from am.scan_address_discovered_by where discovery_id=sga.discovery_id), sga.last_scanned_timestamp, sga.last_seen_timestamp, sga.confidence_score, sga.user_confidence_score, sga.is_soa, sga.is_wildcard_zone, sga.is_hosted_service, sga.ignored, sga.found_from, sga.ns_record, sga.address_hash FROM am.scan_group_addresses as sga WHERE sga.organization_id = $1 AND sga.scan_group_id = $2 AND sga.deleted = $3 AND sga.address_id > $4 ORDER BY sga.address_id LIMIT 1000"
	if query != expected {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}

	filter.Filters.AddBool("ignored", false)
	filter.Filters.AddBool("wildcard", true)
	filter.Filters.AddBool("hosted_service", true)

	filter.Filters.AddInt64("before_scanned_time", time.Now().UnixNano())
	filter.Filters.AddInt64("after_scanned_time", time.Now().UnixNano())

	filter.Filters.AddInt64("before_seen_time", time.Now().UnixNano())
	filter.Filters.AddInt64("after_seen_time", time.Now().UnixNano())

	filter.Filters.AddInt64("before_discovered_time", time.Now().UnixNano())
	filter.Filters.AddInt64("after_discovered_time", time.Now().UnixNano())

	filter.Filters.AddFloat32("above_confidence", 50)
	filter.Filters.AddFloat32("below_confidence", 90)

	filter.Filters.AddFloat32("above_user_confidence", 50)
	filter.Filters.AddFloat32("below_user_confidence", 90)

	filter.Filters.AddInt32("ns_record", 1)
	filter.Filters.AddString("ip_address", "192.168.1.1")
	filter.Filters.AddString("host_address", "example.com")
	filter.Filters.AddString("starts_host_address", "dev")
	filter.Filters.AddString("ends_host_address", "example.com")

	query, args, err = buildGetFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building get query: %v\n", err)
	}
	if len(args) != 21 {
		t.Fatalf("invalid number of args expected 21 got: %d\n", len(args))
	}
	expected = "SELECT sga.organization_id,  sga.address_id,  sga.scan_group_id,  sga.host_address, sga.ip_address,  sga.discovered_timestamp,  (select discovered_by from am.scan_address_discovered_by where discovery_id=sga.discovery_id), sga.last_scanned_timestamp, sga.last_seen_timestamp, sga.confidence_score, sga.user_confidence_score, sga.is_soa, sga.is_wildcard_zone, sga.is_hosted_service, sga.ignored, sga.found_from, sga.ns_record, sga.address_hash FROM am.scan_group_addresses as sga WHERE sga.organization_id = $1 AND sga.scan_group_id = $2 AND sga.deleted = $3 AND sga.ignored = $4 AND sga.is_wildcard_zone = $5 AND sga.last_scanned_timestamp > $6 AND sga.last_scanned_timestamp < $7 AND sga.last_seen_timestamp > $8 AND sga.last_seen_timestamp < $9 AND sga.discovered_timestamp > $10 AND sga.discovered_timestamp < $11 AND sga.confidence_score > $12 AND sga.confidence_score < $13 AND sga.user_confidence_score > $14 AND sga.user_confidence_score < $15 AND sga.ns_record = $16 AND sga.ip_address = $17 AND sga.host_address = $18 AND sga.host_address LIKE $19 AND sga.host_address LIKE $20 AND sga.address_id > $21 ORDER BY sga.address_id LIMIT 1000"
	if query != expected {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}

}
