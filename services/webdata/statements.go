package webdata

import "fmt"

const (
	sharedColumns = `organization_id, 
		address_id, 
		scan_group_id, 
		host_address,
		ip_address, 
		discovered_timestamp, 
		(select discovered_by from am.scan_address_discovered_by where discovery_id=sga.discovery_id),
		last_scanned_timestamp,
		last_seen_timestamp,
		confidence_score,
		user_confidence_score,
		is_soa,
		is_wildcard_zone,
		is_hosted_service,
		ignored,
		found_from,
		ns_record,
		address_hash`
)

var queryMap = map[string]string{
	// am.scan_group_addresses related
	"scanGroupAddressesCount": `select count(address_id) as count from am.scan_group_addresses where organization_id=$1 
		and scan_group_id=$2`,

	// returns
	"scanGroupAddressesAll": fmt.Sprintf(`select 
		%s
		from am.scan_group_addresses as sga where organization_id=$1 and 
		scan_group_id=$2 and 
		address_id > $3 order by address_id limit $4`, sharedColumns),

	"scanGroupAddressesSinceScannedTime": fmt.Sprintf(`select 
		%s
		from am.scan_group_addresses as sga where organization_id=$1 and
		scan_group_id=$2 and 
		(last_scanned_timestamp=0 OR last_scanned_timestamp < $3) and 
		address_id > $4 order by address_id limit $5`, sharedColumns),

	"scanGroupAddressesSinceSeenTime": fmt.Sprintf(`select 
		%s
		from am.scan_group_addresses as sga where organization_id=$1 and
		scan_group_id=$2 and 
		(last_seen_timestamp=0 OR last_seen_timestamp < $3) and 
		address_id > $4 order by address_id limit $5`, sharedColumns),

	"scanGroupAddressesIgnored": fmt.Sprintf(`select 
		%s
		from am.scan_group_addresses as sga where organization_id=$1 and 
		scan_group_id=$2 and 
		ignored=$3 and address_id > $4 order by address_id limit $5`, sharedColumns),
}

var (
	AddResponsesTempTableKey     = "resp_add_temp"
	AddResponsesTempTableColumns = []string{"organization_id", "scan_group_id", "address_id", "response_timestamp",
		"is_document", "scheme", "host", "response_port", "requested_port",
		"url", "headers", "status", "status_text", "mime_type", "data_hash", "raw_data_link"}

	AddResponsesTempTable = `create temporary table resp_add_temp (
			organization_id int,
			scan_group_id int,
			address_id bigint,
			response_timestamp bigint,
			is_document boolean,
			scheme varchar(12),
			host varchar(512),
			response_port int,
			requested_port int,
			url bytea not null,
			headers jsonb,
			status int, 
			status_text text,
			mime_type text,
			data_hash varchar(512),
			raw_data_link text
		) on commit drop;`

	AddResponsesTempToStatus = `insert into am.web_status_text as resp (status_text)
		select temp.status_text from resp_add_temp as temp on conflict do nothing;`

	AddResponsesTempToMime = `insert into am.web_mime_type as resp (mime_type)
		select temp.mime_type from resp_add_temp as temp on conflict do nothing;`

	AddTempToResponses = `insert into am.web_responses as resp (
			organization_id, 
			scan_group_id,
			address_id,
			response_timestamp,
			is_document,
			scheme,
			host,
			response_port,
			requested_port,
			url,
			headers,
			status,
			status_text_id,
			mime_type_id,
			data_hash,
			raw_data_link
		)
		select
			temp.organization_id, 
			temp.scan_group_id, 
			temp.address_id, 
			temp.response_timestamp,
			temp.is_document, 
			temp.scheme,
			temp.host,
			temp.response_port,
			temp.requested_port,
			temp.url,
			temp.headers,
			temp.status,
			(select status_text_id from am.web_status_text where status_text=temp.status_text),
			(select mime_type_id from am.web_mime_type where mime_type=temp.mime_type),
			temp.data_hash,
			temp.raw_data_link,
			false
		from resp_add_temp as temp on conflict do nothing`
)

var (
	AddCertificatesTempTableKey     = "cert_add_temp"
	AddCertificatesTempTableColumns = []string{"organization_id", "scan_group_id", "response_timestamp",
		"host", "port", "protocol", "key_exchange", "key_exchange_group",
		"cipher", "mac", "certificate_value", "subject_name", "san_list", "issuer", "valid_from", "valid_to", "ct_compliance"}

	AddCertificatesTempTable = `create temporary table cert_add_temp (
			organization_id int,
			scan_group_id int,
			response_timestamp bigint,
			host varchar(512),
			port int,
			protocol text,
			key_exchange text,
			key_exchange_group text,
			cipher text,
			mac text,
			certificate_value int,
			subject_name string,
			san_list jsonb,
			issuer string,
			valid_from bigint,
			valid_to bigint,
			ct_compliance 
		) on commit drop;`

	AddTempToCertificates = `insert into am.web_certificates as resp (
			organization_id,
			scan_group_id,
			response_timestamp,
			host,
			port,
			protocol,
			key_exchange,
			key_exchange_group,
			cipher,
			mac,
			certificate_value,
			subject_name,
			san_list,
			issuer,
			valid_from,
			valid_to,
			ct_compliance 
		)
		select
			temp.organization_id,
			temp.scan_group_id,
			temp.response_timestamp,
			temp.host,
			temp.port,
			temp.protocol,
			temp.key_exchange,
			temp.key_exchange_group,
			temp.cipher,
			temp.mac,
			temp.certificate_value,
			temp.subject_name,
			temp.san_list,
			temp.issuer,
			temp.valid_from,
			temp.valid_to,
			temp.ct_compliance, 
			false
		from cert_add_temp as temp on conflict do nothing`
)
