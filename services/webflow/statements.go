package webflow

var resultsColumnsList = `r.web_flow_id,
	r.organization_id,
	r.scan_group_id,
	r.run_timestamp,
	r.url,
	r.load_url,
	r.load_host_address,
	r.load_ip_address,
	r.requested_port,
	r.response_port,
	r.response_timestamp,
	r.result,
	r.response_body_hash,
	r.response_body_link`

var queryMap = map[string]string{
	"getCustomWebScan": `select organization_id, scan_group_id, web_flow_id, web_flow_name, configuration, created_timestamp, modified_timestamp, deleted from am.custom_web_flows where organization_id=$1 and web_flow_id=$2`,

	"customWebScanName": `select organization_id, web_flow_name from am.custom_web_flows where organization_id=$1 and web_flow_id=$2`,

	/*
	   	web_flow_id integer REFERENCES am.custom_web_flows (web_flow_id),
	       organization_id integer REFERENCES am.organizations (organization_id),
	       scan_group_id integer REFERENCES am.scan_group (scan_group_id),
	       last_updated_timestamp timestamptz not null,
	       started_timestamp timestamptz not null,
	       finished_timestamp timestamptz not null,
	       web_flow_status integer not null default 0,
	       total integer not null default 0,
	       in_progress integer not null default 0,
	       completed integer not null default 0
	*/
	"getCustomWebScanStatus": `select organization_id, scan_group_id, last_updated_timestamp, started_timestamp, finished_timestamp, web_flow_status, total, in_progress, completed
		from am.custom_web_flow_status where organization_id=$1 and web_flow_id=$2`,

	"createCustomWebScan": `insert into am.custom_web_flows (organization_id, scan_group_id, web_flow_name, configuration, created_timestamp, modified_timestamp, deleted) values 
		($1, $2, $3, $4, $5, $6, false) returning web_flow_id`,

	"updateCustomWebStatus": `update am.custom_web_flow_status set 
		last_updated_timestamp=now(), total=$1, in_progress=$2, completed=$3 where organization_id=$4 and web_flow_id=$5`,

	"startStopCustomWeb": `update am.custom_web_flow_status set 
	last_updated_timestamp=now(), web_flow_status=$1 where organization_id=$2 and web_flow_id=$3`,

	"deleteCustomWebScan": "update am.custom_web_flows set deleted=true, web_flow_name=$1 where organization_id=$2 and web_flow_id=$3",
}

var (
	AddWebFlowResultsTempTable = `create temporary table result_add_temp (
		web_flow_id integer,
		organization_id integer,
		scan_group_id integer,
		run_timestamp timestamptz not null,
		url bytea not null default '',
		load_url bytea not null default '',
		load_host_address varchar(512) not null default '',
		load_ip_address varchar(256) not null default '',
		requested_port int not null default 0,
		response_port int not null default 0,
		response_timestamp timestamptz not null,
		result jsonb,
		response_body_hash varchar(512) not null default '',
		response_body_link text not null default ''
	) on commit drop;`

	AddWebFlowResultsTempTableKey     = "result_add_temp"
	AddWebFlowResultsTempTableColumns = []string{"web_flow_id", "organization_id", "scan_group_id", "run_timestamp", "url",
		"load_url", "load_host_address", "load_ip_address", "requested_port", "response_port", "response_timestamp", "result", "response_body_hash",
		"response_body_link"}

	AddTempToWebFlowResults = `insert into am.custom_web_flow_results as result (
			web_flow_id,
			organization_id,
			scan_group_id,
			run_timestamp,
			url,
			load_url,
			load_host_address,
			load_ip_address,
			requested_port,
			response_port,
			response_timestamp,
			result,
			response_body_hash,
			response_body_link
		)
		select
			temp.web_flow_id,
			temp.organization_id,
			temp.scan_group_id,
			temp.run_timestamp,
			temp.url,
			temp.load_url,
			temp.load_host_address,
			temp.load_ip_address,
			temp.requested_port,
			temp.response_port,
			temp.response_timestamp,
			temp.result,
			temp.response_body_hash,
			temp.response_body_link
		from result_add_temp as temp on conflict do nothing`
)
