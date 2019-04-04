package webdata

var (
	responseColumns = `wb.response_id,
		organization_id,
		scan_group_id,
		address_hash,
		wb.url_request_timestamp,
		response_timestamp,
		is_document,
		scheme,
		ip_address,
		host_address,
		load_ip_address,
		load_host_address,
		response_port,
		requested_port,
		wb.url,
		headers,
		status, 
		wst.status_text,
		wmt.mime_type,
		raw_body_hash,
		raw_body_link,
		deleted`

	certificateColumns = `certificate_id,
		organization_id,
		scan_group_id,
		response_timestamp,
		address_hash,
		host_address,
		ip_address,
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
		deleted`

	snapshotColumns = `ws.snapshot_id,
		ws.organization_id,
		ws.scan_group_id,
		address_hash,
		host_address,
		ip_address,
		scheme,
		response_port,
		requested_port,
		url,
		response_timestamp,
		serialized_dom_hash,
		serialized_dom_link,
		snapshot_link,
		deleted`

	// do json_agg as opposed to array_agg so nulls aren't horrible to deal with
	techColumns = `json_agg(wtt.category) as category, 
		json_agg(wtt.techname) as techname, 
		json_agg(wt.version) as version, 
		json_agg(wt.match_location) as match_location, 
		json_agg(wt.matched_text) as matched_text, 
		json_agg(wtt.icon) as icon, 
		json_agg(wtt.website) as website`
)

var queryMap = map[string]string{
	"serverCounts": `select 
		scan_group_id,
		agg.server, 
		count(1) as cnt from 
			(select host_address,headers->>'server' as server,
					max(web_responses.url_request_timestamp) as url_request_timestamp,
					max(web_responses.response_timestamp) as response_timestamp from am.web_responses 
					where load_host_address=host_address 
					and load_ip_address=ip_address 
					and organization_id=$1
					group by server,host_address) as agg 
			join am.web_responses as wb 
				on wb.url_request_timestamp=agg.url_request_timestamp 
				and wb.response_timestamp=agg.response_timestamp group by scan_group_id,agg.server order by cnt desc;`,

	"expiringCerts": `select 
		scan_group_id, 'thirty' as days, count(1) from am.web_certificates 
			where (TIMESTAMPTZ 'epoch' + valid_to * '1 second'::interval) 
			between now() and now() + interval '30 days' 
			and organization_id=$1
			group by scan_group_id
		union 
		select scan_group_id, 'fifteen' as days,count(1) from am.web_certificates 
			where (TIMESTAMPTZ 'epoch' + valid_to * '1 second'::interval) 
			between now() and now() + interval '15 days'
			and organization_id=$1
			group by scan_group_id`,

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

	"insertSnapshot": `insert into am.web_snapshots (organization_id, scan_group_id, address_hash, host_address, ip_address, scheme, response_port, requested_port, url, response_timestamp, serialized_dom_hash, serialized_dom_link, snapshot_link, deleted)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, false) 
		on conflict (organization_id, scan_group_id, address_hash, serialized_dom_hash, response_port) do update set
			response_timestamp=EXCLUDED.response_timestamp
			returning snapshot_id`,

	"insertWebTech": `insert into am.web_technologies (snapshot_id, organization_id, scan_group_id, techtype_id, matched_text, match_location, version)
		(select $1, $2, $3, techtype_id, $4, $5, $6 from am.web_techtypes where techname=$7) on conflict (snapshot_id,techtype_id,match_location) do update set
			matched_text=EXCLUDED.matched_text,
			match_location=EXCLUDED.match_location,
			version=EXCLUDED.version`,
}

var (
	AddResponsesTempTableKey     = "resp_add_temp"
	AddResponsesTempTableColumns = []string{"organization_id", "scan_group_id", "address_hash", "url_request_timestamp", "response_timestamp",
		"is_document", "scheme", "ip_address", "host_address", "load_ip_address", "load_host_address", "response_port", "requested_port",
		"url", "headers", "status", "status_text", "mime_type", "raw_body_hash", "raw_body_link"}

	AddResponsesTempTable = `create temporary table resp_add_temp (
			organization_id int,
			scan_group_id int,
			address_hash varchar(128),
			url_request_timestamp timestamptz,
			response_timestamp timestamptz,
			is_document boolean,
			scheme varchar(12),
			ip_address varchar(256),
			host_address varchar(512),
			load_ip_address varchar(256),
			load_host_address varchar(512),
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
			address_hash,
			url_request_timestamp,
			response_timestamp,
			is_document,
			scheme,
			ip_address,
			host_address,
			load_ip_address,
			load_host_address,
			response_port,
			requested_port,
			url,
			headers,
			status,
			status_text_id,
			mime_type_id,
			raw_body_hash,
			raw_body_link,
			deleted
		)
		select
			temp.organization_id, 
			temp.scan_group_id, 
			temp.address_hash, 
			temp.url_request_timestamp,
			temp.response_timestamp,
			temp.is_document, 
			temp.scheme,
			temp.ip_address,
			temp.host_address,
			temp.load_ip_address,
			temp.load_host_address,
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
		"address_hash", "host_address", "ip_address", "port", "protocol", "key_exchange", "key_exchange_group",
		"cipher", "mac", "certificate_value", "subject_name", "san_list", "issuer", "valid_from", "valid_to", "ct_compliance"}

	AddCertificatesTempTable = `create temporary table cert_add_temp (
			organization_id int,
			scan_group_id int,
			response_timestamp timestamptz,
			address_hash varchar(128),
			host_address varchar(512),
			ip_address varchar(256),
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
			address_hash,
			host_address,
			ip_address,
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
			deleted 
		)
		select
			temp.organization_id,
			temp.scan_group_id,
			temp.response_timestamp,
			temp.address_hash,
			temp.host_address,
			temp.ip_address,
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
			valid_from=EXCLUDED.valid_from,
			valid_to=EXCLUDED.valid_to,
			response_timestamp=EXCLUDED.response_timestamp`
)
