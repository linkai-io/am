package webflow

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

	"updateCustomWebStatus": `insert into am.custom_web_flow_status (web_flow_id, organization_id, scan_group_id, last_updated_timestamp, 
		started_timestamp, finished_timestamp, web_flow_status, total, in_progress, completed) values 
		($1, $2, $3, 44, $5, $6, $7, $8, $9, $10)`,

	"deleteCustomWebScan": "update am.custom_web_flows set deleted=true, web_flow_name=$1 where organization_id=$2 and web_flow_id=$3",
	"startCustomWebScan":  "",
	"stopCustomWebScan":   "",
}
