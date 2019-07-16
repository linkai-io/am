package address

import "fmt"

const (
	sharedColumns = `sga.organization_id, 
		sga.address_id, 
		sga.scan_group_id, 
		sga.host_address,
		sga.ip_address, 
		sga.discovered_timestamp, 
		(select discovered_by from am.scan_address_discovered_by where discovery_id=sga.discovery_id),
		sga.last_scanned_timestamp,
		sga.last_seen_timestamp,
		sga.confidence_score,
		sga.user_confidence_score,
		sga.is_soa,
		sga.is_wildcard_zone,
		sga.is_hosted_service,
		sga.ignored,
		sga.found_from,
		sga.ns_record,
		sga.address_hash`

	defaultColumns = `address_id, organization_id, scan_group_id, host_address, ip_address, 
		discovery_id, confidence_score, user_confidence_score, is_soa, is_wildcard_zone, is_hosted_service, 
		ignored, found_from, ns_record, address_hash, discovered_timestamp, last_scanned_timestamp, last_seen_timestamp, deleted`

	defaultPortColumns = `port_id, organization_id, scan_group_id, host_address, port_data, scanned_timestamp, previous_scanned_timestamp, is_ipv4`
)

var queryMap = map[string]string{
	// aggregates
	"discoveredOrgAgg": `select 'discovery_day' as agg,scan_group_id, period_start, sum(discovered_count) as cnt FROM am.discoveries_1day
	WHERE organization_id=$1 and period_start > $2 GROUP BY scan_group_id, period_start, discovered_count
union select 'seen_day' as agg,scan_group_id, period_start, sum(seen_count) as cnt FROM am.seen_1day
	WHERE organization_id=$1 and period_start > $2 GROUP BY scan_group_id, period_start, seen_count
union select 'scanned_day' as agg,scan_group_id, period_start, sum(scanned_count) as cnt FROM am.scanned_1day
	WHERE organization_id=$1 and period_start > $2 GROUP BY scan_group_id, period_start, scanned_count 
union select 'discovery_trihourly' as agg,scan_group_id, period_start, sum(discovered_count) as cnt FROM am.discoveries_3hour
	WHERE organization_id=$1 and period_start > $2 GROUP BY scan_group_id, period_start, discovered_count
union select 'seen_trihourly' as agg,scan_group_id, period_start, sum(seen_count) as cnt FROM am.seen_3hour
	WHERE organization_id=$1 and period_start > $2 GROUP BY scan_group_id, period_start, seen_count
union select 'scanned_trihourly' as agg,scan_group_id, period_start, sum(scanned_count) as cnt FROM am.scanned_3hour
	WHERE organization_id=$1 and period_start > $2 GROUP BY scan_group_id, period_start, scanned_count 
	order by period_start desc;`,

	"discoveredByOrg": `select 
		scan_group_id, 
		(select discovered_by from am.scan_address_discovered_by where discovery_id=sga.discovery_id) as discovered_by, 
		count(1) as total from am.scan_group_addresses as sga 
			where organization_id=$1 
			and confidence_score=100 or user_confidence_score=100 
			and ignored=false 
			and deleted=false
			group by scan_group_id,discovered_by`,

	"discoveredByGroup": `select 
		(select discovered_by from am.scan_address_discovered_by where discovery_id=sga.discovery_id) as discovered_by, 
		count(1) as total from am.scan_group_addresses as sga 
			where sga.organization_id=$1 
			and sga.scan_group_id=$2
			and sga.confidence_score=100 or sga.user_confidence_score=100 
			and sga.ignored=false 
			and sga.deleted=false
			group by sga.scan_group_id,discovered_by`,

	// am.scan_group_addresses related
	"scanGroupAddressesCount": `select count(address_id) as count from am.scan_group_addresses where organization_id=$1 
		and scan_group_id=$2 and deleted=false`,

	"scanGroupHostList": `select 
			top.organization_id, 
			top.scan_group_id, 
			top.host_address, 
			array_agg(arr.ip_address) as addresses, 
			array_agg(arr.address_id) as address_ids,
			ports.scanned_timestamp,
			ports.previous_scanned_timestamp,
			ports.port_data 
		from am.scan_group_addresses as top 
			left join am.scan_group_addresses as arr on top.address_id=arr.address_id 
			left join am.scan_group_addresses_ports as ports on top.host_address=ports.host_address and top.organization_id=ports.organization_id and top.scan_group_id=ports.scan_group_id
		where top.organization_id=$1 and top.scan_group_id=$2 
			and top.host_address != '' 
			and top.deleted=false
			and top.ignored=false
			and (top.confidence_score=100 or top.user_confidence_score=100)
			and top.host_address > $3 
			group by top.organization_id, top.scan_group_id, top.host_address, ports.scanned_timestamp,
			ports.previous_scanned_timestamp, ports.port_data order by top.host_address limit $4;`,

	"unsetMaxHosts": `update am.organizations set limit_hosts_reached=false where organization_id=$1`,

	"archiveHosts": fmt.Sprintf(`with archive_time as (
		select now() as archived_timestamp
	),
	delete_ports as (
		delete from am.scan_group_addresses_ports where 
		organization_id=$1 and scan_group_id=$2 and scanned_timestamp < $3 returning %s
	),
	insert_ports as ( 
		insert into am.scan_group_addresses_ports_archive select %s,archive_time.archived_timestamp from delete_ports,archive_time
	),	
	delete_hosts as (
		delete from am.scan_group_addresses where 
		organization_id=$1 and scan_group_id=$2 and discovery_id not in (1,2) and last_seen_timestamp < $3 returning %s
	) 
	insert into am.scan_group_addresses_archive (%s, archived_timestamp) select %s,archive_time.archived_timestamp from delete_hosts,archive_time`, defaultPortColumns, defaultPortColumns, defaultColumns, defaultColumns, defaultColumns),

	// PORT SCAN STMTS
	"insertPortHost": `insert into am.scan_group_addresses (organization_id, scan_group_id, host_address, ip_address, 
		discovery_id, confidence_score, user_confidence_score, is_soa, is_wildcard_zone, is_hosted_service, 
		ignored, found_from, ns_record, address_hash, discovered_timestamp, last_scanned_timestamp, last_seen_timestamp) 
		values ($1, $2, $3, $4, (select discovery_id from am.scan_address_discovered_by where discovered_by=$5), $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) on conflict do nothing;`,

	"updateHostPorts": `insert into am.scan_group_addresses_ports (organization_id, scan_group_id, host_address, port_data, scanned_timestamp) 
		values ($1, $2, $3, $4, $5) on conflict (organization_id, scan_group_id, host_address, is_ipv4) do update set
		previous_scanned_timestamp=am.scan_group_addresses_ports.scanned_timestamp,
		scanned_timestamp=EXCLUDED.scanned_timestamp,
		port_data = EXCLUDED.port_data || ( 
			jsonb_set(EXCLUDED.port_data, '{previous}', am.scan_group_addresses_ports.port_data->'current', true)
		)
		`,
}

var (
	AddAddressesTempTableKey     = "sga_add_temp"
	AddAddressesTempTableColumns = []string{"organization_id", "scan_group_id", "host_address", "ip_address",
		"discovered_timestamp", "discovered_by", "last_scanned_timestamp", "last_seen_timestamp", "confidence_score",
		"user_confidence_score", "is_soa", "is_wildcard_zone", "is_hosted_service", "ignored", "found_from", "ns_record", "address_hash"}
	AddAddressesTempTable = `create temporary table sga_add_temp (
			organization_id integer not null,
			scan_group_id integer not null,
			host_address varchar(512),
			ip_address varchar(256),
			discovered_timestamp timestamptz,
			discovered_by varchar,
			last_scanned_timestamp timestamptz,
			last_seen_timestamp timestamptz,
			confidence_score float,
			user_confidence_score float,
			is_soa boolean not null,
			is_wildcard_zone boolean not null,
			is_hosted_service boolean not null,
			ignored boolean not null,
			found_from varchar(128),
			ns_record int,
			address_hash varchar(128)
			check (host_address is not null or ip_address is not null)
		) on commit drop;`

	AddAddressesTempToAddress = `insert into am.scan_group_addresses as sga (
			organization_id, 
			scan_group_id,
			host_address,
			ip_address,
			discovered_timestamp,
			discovery_id,
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
			address_hash
		)
		select
			temp.organization_id, 
			temp.scan_group_id, 
			temp.host_address, 
			temp.ip_address,
			temp.discovered_timestamp, 
			(select discovery_id from am.scan_address_discovered_by where discovered_by=temp.discovered_by),
			temp.last_scanned_timestamp,
			temp.last_seen_timestamp,
			temp.confidence_score,
			temp.user_confidence_score,
			temp.is_soa,
			temp.is_wildcard_zone,
			temp.is_hosted_service,
			temp.ignored,
			temp.found_from,
			temp.ns_record,
			temp.address_hash
		from sga_add_temp as temp on conflict (scan_group_id, address_hash) do update set
			last_scanned_timestamp=EXCLUDED.last_scanned_timestamp,
			last_seen_timestamp=EXCLUDED.last_seen_timestamp,
			confidence_score=EXCLUDED.confidence_score,
			user_confidence_score=EXCLUDED.user_confidence_score,
			is_soa=EXCLUDED.is_soa,
			is_wildcard_zone=EXCLUDED.is_wildcard_zone,
			is_hosted_service=EXCLUDED.is_hosted_service,
			ignored=EXCLUDED.ignored`

	DeleteAddressesTempTableKey     = "sga_del_temp"
	DeleteAddressesTempTableColumns = []string{"address_id"}
	DeleteAddressesTempTable        = `create temporary table sga_del_temp (
			address_id bigint not null
		) on commit drop;`

	DeleteAddressesTempToAddress = `update am.scan_group_addresses as sga 
		set deleted=true 
		from sga_del_temp as temp
		where sga.address_id=temp.address_id and sga.organization_id=$1 and sga.scan_group_id=$2`

	IgnoreAddressesTempTableKey     = "sga_ignore_temp"
	IgnoreAddressesTempTableColumns = []string{"address_id"}
	IgnoreAddressesTempTable        = `create temporary table sga_ignore_temp (
			address_id bigint not null
		) on commit drop;`

	IgnoreAddressesTempToAddress = `update am.scan_group_addresses as sga
		set ignored=$1 
		from sga_ignore_temp as temp
		where sga.address_id=temp.address_id and sga.organization_id=$2 and sga.scan_group_id=$3`
)
