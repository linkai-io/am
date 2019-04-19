package webflow

/*
	organization_id integer REFERENCES am.organizations (organization_id),
    scan_group_id integer REFERENCES am.scan_group (scan_group_id),
    web_flow_name varchar(128) not null default '',
    configuration jsonb,
    created_timestamp timestamptz not null,
    modified_timestamp timestamptz not null,
	deleted boolean default false,
*/
var queryMap = map[string]string{
	"getCustomWebScan":  `select organization_id, scan_group_id, web_flow_name, configuration, created_timestamp, modified_timestamp, deleted from am.custom_web_flows where organization_id=$1 and web_flow_id=$2`,
	"customWebScanName": `select organization_id, web_flow_name from am.custom_web_flows where organization_id=$1 and web_flow_id=$2`,
	"createCustomWebScan": `insert into am.custom_web_flows (organization_id, scan_group_id, web_flow_name, configuration, created_timestamp, modified_timestamp, deleted) values 
		($1, $2, $3, $4, $5, $6, false) returning web_flow_id`,
	"deleteCustomWebScan": "update am.custom_web_flows set deleted=true, web_flow_name=$1 where organization_id=$2 and web_flow_id=$3",
	"startCustomWebScan":  "",
	"stopCustomWebScan":   "",
}
