package webdata

import (
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
)

func TestBuildURLListFilterQuery(t *testing.T) {
	userContext := amtest.CreateUserContext(1, 1)
	filter := &am.WebResponseFilter{
		OrgID:   1,
		GroupID: 1,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}

	query, args, err := buildURLListFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building query %v\n", err)
	}

	t.Logf("%s\n", query)
	t.Logf("%#v\n", args)
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d\n", len(args))
	}
	expected := "SELECT wb.organization_id, wb.scan_group_id, wb.url_request_timestamp, load_host_address, load_ip_address, array_agg(wb.url) as urls, array_agg(wb.raw_body_link) as raw_body_links, array_agg(wb.response_id) as response_ids, array_agg((select mime_type from am.web_mime_type where mime_type_id=wb.mime_type_id)) as mime_types FROM am.web_responses as wb WHERE wb.organization_id = $1 AND wb.scan_group_id = $2 AND wb.response_id > $3 GROUP BY wb.organization_id, wb.scan_group_id, load_host_address, load_ip_address, wb.url_request_timestamp ORDER BY wb.url_request_timestamp LIMIT 1000"
	if expected != query {
		t.Fatalf("Expected:\n%v\nGot:\n%v\n", expected, query)
	}

	// test after response time
	filter.Filters.AddInt64("after_request_time", time.Now().UnixNano())
	query, args, err = buildURLListFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error after_request_time building query %v\n", err)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}
}

func TestBuildURLListFilterQueryLatestOnly(t *testing.T) {
	userContext := amtest.CreateUserContext(1, 1)
	filter := &am.WebResponseFilter{
		OrgID:   1,
		GroupID: 1,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}

	filter.Filters.AddBool("latest_only", true)
	query, args, err := buildURLListFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building query %v\n", err)
	}

	t.Logf("%s\n", query)
	t.Logf("%#v\n", args)
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d\n", len(args))
	}
	expected := "SELECT wb.organization_id, wb.scan_group_id, wb.url_request_timestamp, load_host_address, load_ip_address, array_agg(wb.url) as urls, array_agg(wb.raw_body_link) as raw_body_links, array_agg(wb.response_id) as response_ids, array_agg((select mime_type from am.web_mime_type where mime_type_id=wb.mime_type_id)) as mime_types FROM (SELECT url, (max(url_request_timestamp)) AS url_request_timestamp FROM am.web_responses GROUP BY url) AS latest JOIN am.web_responses as wb on wb.url=latest.url and wb.url_request_timestamp=latest.url_request_timestamp WHERE wb.organization_id = $1 AND wb.scan_group_id = $2 AND wb.response_id > $3 GROUP BY wb.organization_id, wb.scan_group_id, load_host_address, load_ip_address, wb.url_request_timestamp ORDER BY wb.url_request_timestamp LIMIT 1000"
	if expected != query {
		t.Fatalf("Expected:\n%v\nGot:\n%v\n", expected, query)
	}

	// test after response time
	filter.Filters.AddInt64("after_request_time", time.Now().UnixNano())
	query, args, err = buildURLListFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error after_request_time building query %v\n", err)
	}
	t.Logf("%s\n", query)
	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}
}

func TestBuildWebFilterQuery(t *testing.T) {
	userContext := amtest.CreateUserContext(1, 1)
	filter := &am.WebResponseFilter{
		OrgID:   1,
		GroupID: 1,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}

	query, args, err := buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d\n", len(args))
	}

	expected := "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM am.web_responses as wb JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 LIMIT 1000"
	if query != expected {
		t.Fatalf("Expected:\n%v\nGot:\n:%v", expected, query)
	}

	filter.Filters.AddStrings("mime_type", []string{"text/html", "application/json"})
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM am.web_responses as wb JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 AND mime_type IN ($4,$5) LIMIT 1000"
	if query != expected {
		t.Fatalf("Expected:\n%v\nGot:\n:%v", expected, query)
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddInt64("after_request_time", time.Now().UnixNano())
	filter.Filters.AddInt64("before_request_time", time.Now().UnixNano())
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM am.web_responses as wb JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 AND wb.url_request_timestamp > $4 AND wb.url_request_timestamp < $5 LIMIT 1000"
	if expected != query {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddString("header_names", "x-content-type")
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM am.web_responses as wb JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 AND headers ? $4 LIMIT 1000"
	if expected != query {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}

	filter.Filters.AddString("header_names", "content-length")
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM am.web_responses as wb JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 AND headers ? $4 AND headers ? $5 LIMIT 1000"
	if expected != query {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddString("header_pair_names", "server")
	filter.Filters.AddString("header_pair_values", "AmazonS3")
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM am.web_responses as wb JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 AND headers->>$4=$5 LIMIT 1000"
	if expected != query {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}
}

func TestBuildWebFilterLatestOnlyQuery(t *testing.T) {
	userContext := amtest.CreateUserContext(1, 1)
	filter := &am.WebResponseFilter{
		OrgID:   1,
		GroupID: 1,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	filter.Filters.AddBool("latest_only", true)
	query, args, err := buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d\n", len(args))
	}

	expected := "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM (SELECT web_responses.url, (max(web_responses.url_request_timestamp)) AS url_request_timestamp FROM am.web_responses WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 GROUP BY url) AS latest JOIN am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp and wb.url=latest.url JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id ORDER BY response_id LIMIT 1000"
	if query != expected {
		t.Fatalf("Expected:\n%v\nGot:\n%v", expected, query)
	}

	filter.Filters.AddStrings("mime_type", []string{"text/html", "application/json"})
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM (SELECT web_responses.url, (max(web_responses.url_request_timestamp)) AS url_request_timestamp FROM am.web_responses WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 GROUP BY url) AS latest JOIN am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp and wb.url=latest.url JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE mime_type IN ($4,$5) ORDER BY response_id LIMIT 1000"
	if query != expected {
		t.Fatalf("Expected:\n%v\nGot:\n%v", expected, query)
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddBool("latest_only", true)
	filter.Filters.AddInt64("after_request_time", time.Now().UnixNano())
	filter.Filters.AddInt64("before_request_time", time.Now().UnixNano())
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM (SELECT web_responses.url, (max(web_responses.url_request_timestamp)) AS url_request_timestamp FROM am.web_responses WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 GROUP BY url) AS latest JOIN am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp and wb.url=latest.url JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE wb.url_request_timestamp > $4 AND wb.url_request_timestamp < $5 ORDER BY response_id LIMIT 1000"
	if expected != query {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddBool("latest_only", true)
	filter.Filters.AddString("header_names", "x-content-type")
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM (SELECT web_responses.url, (max(web_responses.url_request_timestamp)) AS url_request_timestamp FROM am.web_responses WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 GROUP BY url) AS latest JOIN am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp and wb.url=latest.url JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE headers ? $4 ORDER BY response_id LIMIT 1000"
	if expected != query {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}

	filter.Filters.AddString("header_names", "content-length")
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM (SELECT web_responses.url, (max(web_responses.url_request_timestamp)) AS url_request_timestamp FROM am.web_responses WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 GROUP BY url) AS latest JOIN am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp and wb.url=latest.url JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE headers ? $4 AND headers ? $5 ORDER BY response_id LIMIT 1000"
	if expected != query {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddBool("latest_only", true)
	filter.Filters.AddString("header_pair_names", "server")
	filter.Filters.AddString("header_pair_values", "AmazonS3")
	query, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error building web filter: %v\n", err)
	}

	t.Logf("%#v\n", query)
	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d\n", len(args))
	}

	expected = "SELECT wb.response_id, organization_id, scan_group_id, address_hash, wb.url_request_timestamp, response_timestamp, is_document, scheme, ip_address, host_address, load_ip_address, load_host_address, response_port, requested_port, wb.url, headers, status,  wst.status_text, wmt.mime_type, raw_body_hash, raw_body_link, deleted FROM (SELECT web_responses.url, (max(web_responses.url_request_timestamp)) AS url_request_timestamp FROM am.web_responses WHERE organization_id = $1 AND scan_group_id = $2 AND response_id > $3 GROUP BY url) AS latest JOIN am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp and wb.url=latest.url JOIN am.web_status_text as wst on wb.status_text_id = wst.status_text_id JOIN am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id WHERE headers->>$4=$5 ORDER BY response_id LIMIT 1000"
	if expected != query {
		t.Fatalf("expected:\n%v\ngot:\n%v\n", expected, query)
	}
}
