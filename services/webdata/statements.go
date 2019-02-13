package webdata

import "fmt"

var (
	responseColumns = `wb.response_id,
		organization_id,
		scan_group_id,
		address_id,
		wb.url_request_timestamp,
		response_timestamp,
		is_document,
		scheme,
		ip_address,
		host_address,
		response_port,
		requested_port,
		wb.url,
		headers,
		status, 
		wst.status_text,
		wmt.mime_type,
		raw_body_hash,
		raw_body_link,
		is_deleted`

	// for adding additional context to responses
	referencedResponseColumns = `(select host_address from am.scan_group_addresses as sga where sga.address_id=wb.address_id) as address_id_host_address,
	(select ip_address from am.scan_group_addresses as sga where sga.address_id=wb.address_id) as address_id_ip_address`

	responseQueryPrefix = fmt.Sprintf(`select %s,%s from am.web_responses as wb 
	join am.web_status_text as wst on wb.status_text_id = wst.status_text_id
	join am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id
		where organization_id=$1 and
		scan_group_id=$2 and 
		`, responseColumns, referencedResponseColumns)

	latestOnlyResponseQueryPrefix = fmt.Sprintf(`select distinct %s,
		%s from (select web_responses.url, max(web_responses.url_request_timestamp) as url_request_timestamp from am.web_responses 
			where organization_id=$1 and
			scan_group_id=$2 and
		`, responseColumns, referencedResponseColumns)

	latestOnlyUrlListQueryPrefix = `select wb.organization_id,
	wb.scan_group_id,
	latest.url_request_timestamp,
	(select host_address from am.scan_group_addresses as sga where sga.address_id=wb.address_id) as address_id_host_address,
	(select ip_address from am.scan_group_addresses as sga where sga.address_id=wb.address_id) as address_id_ip_address,
	array_agg(wb.url),
	array_agg(wb.raw_body_link) as raw_body_links,
	array_agg(wb.response_id) as response_ids,
	array_agg((select mime_type from am.web_mime_type where wb.mime_type_id=wmt.mime_type_id)) as mime_types
	 from (select url, max(url_request_timestamp) as url_request_timestamp from am.web_responses group by url) as latest
	join am.web_responses as wb on wb.url=latest.url and wb.url_request_timestamp=latest.url_request_timestamp
	join am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id
	where wb.organization_id=$1 and wb.scan_group_id=$2`

	urlListQueryPrefix = `select wb.organization_id,
		wb.scan_group_id,
		wb.url_request_timestamp,
		(select host_address from am.scan_group_addresses as sga where sga.address_id=wb.address_id) as address_id_host_address,
		(select ip_address from am.scan_group_addresses as sga where sga.address_id=wb.address_id) as address_id_ip_address,
		array_agg(wb.url),
		array_agg(wb.raw_body_link) as raw_body_links,
		array_agg(wb.response_id) as response_ids,
		array_agg((select mime_type from am.web_mime_type where wb.mime_type_id=wmt.mime_type_id)) as mime_types
		from am.web_responses as wb
		join am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id
		where wb.organization_id=$1 and scan_group_id=$2 `

	certificateColumns = `certificate_id,
		organization_id,
		scan_group_id,
		response_timestamp,
		host_address,
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
		ct_compliance,
		is_deleted`

	snapshotColumns = `snapshot_id,
		organization_id,
		scan_group_id,
		address_id,
		response_timestamp,
		serialized_dom_hash,
		serialized_dom_link,
		snapshot_link,
		is_deleted`

	// for adding additional context to snap shots
	referencedSnapshotColumns = `(select host_address from am.scan_group_addresses as sga where sga.address_id=ws.address_id) as address_id_host_address,
	(select ip_address from am.scan_group_addresses as sga where sga.address_id=ws.address_id) as address_id_ip_address`
)

var queryMap = map[string]string{
	"responseURLList": `select 
		top.organization_id, 
		top.scan_group_id, 
		top.host_address, 
		array_agg(arr.ip_address) as addresses, 
		array_agg(arr.address_id) as address_ids 
			from am.scan_group_addresses as top 
		left join am.scan_group_addresses as arr on 
			top.address_id=arr.address_id 
		where top.organization_id=$1 and top.scan_group_id=$2 and top.host_address != '' group by top.organization_id, top.scan_group_id, top.host_address;`,

	"insertSnapshot": `insert into am.web_snapshots (organization_id, scan_group_id, address_id, response_timestamp, serialized_dom_hash, serialized_dom_link, snapshot_link, is_deleted)
			values ($1, $2, $3, $4, $5, $6, $7, false) 
		on conflict (organization_id, scan_group_id, serialized_dom_hash) do update set
			response_timestamp=EXCLUDED.response_timestamp`,

	"certificatesSinceResponseTime": fmt.Sprintf(`select %s from am.web_certificates as wb 
		where organization_id=$1 and
		scan_group_id=$2 and 
		response_timestamp > $3 and 
		is_deleted = false and
		certificate_id > $4 order by certificate_id limit $5`, certificateColumns),

	"certificatesAll": fmt.Sprintf(`select %s from am.web_certificates as wc 
		where organization_id=$1 and
		scan_group_id=$2 and 
		is_deleted = false and
		certificate_id > $3 order by certificate_id limit $4`, certificateColumns),

	"snapshotsSinceResponseTime": fmt.Sprintf(`select %s,%s from am.web_snapshots as ws 
		where organization_id=$1 and
		scan_group_id=$2 and 
		response_timestamp > $3 and 
		is_deleted = false and
		snapshot_id > $4 order by snapshot_id limit $5`, snapshotColumns, referencedSnapshotColumns),

	"snapshotsAll": fmt.Sprintf(`select %s,%s from am.web_snapshots as ws 
		where organization_id=$1 and
		scan_group_id=$2 and 
		is_deleted = false and
		snapshot_id > $3 order by snapshot_id limit $4`, snapshotColumns, referencedSnapshotColumns),
}

var (
	AddResponsesTempTableKey     = "resp_add_temp"
	AddResponsesTempTableColumns = []string{"organization_id", "scan_group_id", "address_id", "url_request_timestamp", "response_timestamp",
		"is_document", "scheme", "ip_address", "host_address", "response_port", "requested_port",
		"url", "headers", "status", "status_text", "mime_type", "raw_body_hash", "raw_body_link"}

	AddResponsesTempTable = `create temporary table resp_add_temp (
			organization_id int,
			scan_group_id int,
			address_id bigint,
			url_request_timestamp bigint,
			response_timestamp bigint,
			is_document boolean,
			scheme varchar(12),
			ip_address varchar(256),
			host_address varchar(512),
			response_port int,
			requested_port int,
			url bytea not null,
			headers jsonb,
			status int, 
			status_text text,
			mime_type text,
			raw_body_hash varchar(512),
			raw_body_link text
		) on commit drop;`

	AddResponsesTempToStatus = `insert into am.web_status_text as resp (status_text)
		select temp.status_text from resp_add_temp as temp on conflict do nothing;`

	AddResponsesTempToMime = `insert into am.web_mime_type as resp (mime_type)
		select temp.mime_type from resp_add_temp as temp on conflict do nothing;`

	AddTempToResponses = `insert into am.web_responses as resp (
			organization_id, 
			scan_group_id,
			address_id,
			url_request_timestamp,
			response_timestamp,
			is_document,
			scheme,
			ip_address,
			host_address,
			response_port,
			requested_port,
			url,
			headers,
			status,
			status_text_id,
			mime_type_id,
			raw_body_hash,
			raw_body_link,
			is_deleted
		)
		select
			temp.organization_id, 
			temp.scan_group_id, 
			temp.address_id, 
			temp.url_request_timestamp,
			temp.response_timestamp,
			temp.is_document, 
			temp.scheme,
			temp.ip_address,
			temp.host_address,
			temp.response_port,
			temp.requested_port,
			temp.url,
			temp.headers,
			temp.status,
			(select status_text_id from am.web_status_text where status_text=temp.status_text),
			(select mime_type_id from am.web_mime_type where mime_type=temp.mime_type),
			temp.raw_body_hash,
			temp.raw_body_link,
			false
		from resp_add_temp as temp on conflict do nothing`
)

var (
	AddCertificatesTempTableKey     = "cert_add_temp"
	AddCertificatesTempTableColumns = []string{"organization_id", "scan_group_id", "response_timestamp",
		"host_address", "port", "protocol", "key_exchange", "key_exchange_group",
		"cipher", "mac", "certificate_value", "subject_name", "san_list", "issuer", "valid_from", "valid_to", "ct_compliance"}

	AddCertificatesTempTable = `create temporary table cert_add_temp (
			organization_id int,
			scan_group_id int,
			response_timestamp bigint,
			host_address varchar(512),
			port int,
			protocol text,
			key_exchange text,
			key_exchange_group text,
			cipher text,
			mac text,
			certificate_value int,
			subject_name varchar(512),
			san_list jsonb,
			issuer text,
			valid_from bigint,
			valid_to bigint,
			ct_compliance text 
		) on commit drop;`

	AddTempToCertificates = `insert into am.web_certificates as certs (
			organization_id,
			scan_group_id,
			response_timestamp,
			host_address,
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
			ct_compliance,
			is_deleted 
		)
		select
			temp.organization_id,
			temp.scan_group_id,
			temp.response_timestamp,
			temp.host_address,
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
		from cert_add_temp as temp on conflict (organization_id, scan_group_id, subject_name, valid_from, valid_to) do update set
			response_timestamp=EXCLUDED.response_timestamp`
)
