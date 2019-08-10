package event

import "fmt"

const sharedSettingColumns = `organization_id, 
	user_id, 
	weekly_report_send_day, 
	daily_report_send_hour, 
	user_timezone, 
	should_weekly_email,
	should_daily_email`

var queryMap = map[string]string{
	"getUserSettings":      fmt.Sprintf(`select %s from am.user_notification_settings where organization_id=$1 and user_id=$2`, sharedSettingColumns),
	"getUserSubscriptions": `select organization_id, user_id, type_id, subscribed_since, subscribed, webhook_version, webhook_enabled, webhook_url, webhook_type from am.user_notification_subscriptions where organization_id=$1 and user_id=$2`,
	"updateUserSettings": fmt.Sprintf(`insert into am.user_notification_settings (%s) values
		($1, $2, $3, $4, $5, $6, $7) on conflict (organization_id,user_id) do update set
		weekly_report_send_day=EXCLUDED.weekly_report_send_day,
		daily_report_send_hour=EXCLUDED.daily_report_send_hour,
		user_timezone=EXCLUDED.user_timezone,
		should_weekly_email=EXCLUDED.should_weekly_email,
		should_daily_email=EXCLUDED.should_daily_email`, sharedSettingColumns),

	"newHostnames": `select latest.host_address from am.scan_group_addresses as latest 
		where organization_id=$1 and scan_group_id=$2 and (confidence_score=100 or user_confidence_score=100) and
		deleted=false and
		ignored=false and discovered_timestamp >= $3
		and not exists (
		select sga.host_address from am.scan_group_addresses as sga
			where sga.discovered_timestamp <= $3 and
			organization_id=$1 and scan_group_id=$2 and (confidence_score=100 or user_confidence_score=100) and
			deleted=false and
			ignored=false group by sga.host_address
		)
		group by latest.host_address`,

	"newWebsites": `select latest.load_url, latest.url, latest.response_port from am.web_snapshots as latest
		where deleted=false and 
		updated=false and 
		organization_id=$1 and 
		scan_group_id=$2 and 
		url_request_timestamp >= $3
		and not exists (
			select load_url, response_port from am.web_snapshots as ws 
			where ws.url_request_timestamp < $3 and ws.organization_id=$1 and ws.scan_group_id=$2 and 
			latest.load_url=ws.load_url and latest.response_port=ws.response_port 
			group by ws.load_url,  ws.response_port
		)
		group by latest.load_url, latest.url, latest.response_port;`,

	"newTechnologies": `with prev_tech as (
		select ws.load_url, ws.url, ws.response_port, wtt.techname, t.version from am.web_technologies as t
		join am.web_techtypes as wtt on t.techtype_id=wtt.techtype_id
		join am.web_snapshots as ws on t.snapshot_id=ws.snapshot_id
		where ws.organization_id=$1 and ws.scan_group_id=$2 
		and	ws.url_request_timestamp < $3 
		group by ws.load_url, ws.url, ws.response_port, wtt.techname, t.version
	),
	new_tech as (
		select ws.load_url, ws.url, ws.response_port, wtt.techname, t.version from am.web_technologies as t
		join am.web_snapshots as ws on t.snapshot_id=ws.snapshot_id
		join am.web_techtypes as wtt on t.techtype_id=wtt.techtype_id
		where ws.organization_id=$1 and ws.scan_group_id=$2
		and ws.url_request_timestamp >= $3 
		and t.updated<>true
		group by ws.load_url, ws.url, ws.response_port, wtt.techname, t.version
	)
	select new_tech.load_url, new_tech.url, new_tech.response_port, new_tech.techname, new_tech.version from new_tech where not exists (
		select prev_tech.load_url, prev_tech.techname,prev_tech.version from prev_tech where 
		new_tech.load_url=prev_tech.load_url and new_tech.response_port=prev_tech.response_port and new_tech.techname=prev_tech.techname and new_tech.version=prev_tech.version
	)`,

	"checkPortChanges": `select host_address, port_data, scanned_timestamp, previous_scanned_timestamp from am.scan_group_addresses_ports
		where organization_id=$1
		and scan_group_id=$2
		and scanned_timestamp >= $3
		and previous_scanned_timestamp != 'epoch'`,

	"checkCertExpiration": `select subject_name, port, valid_to from am.web_certificates 
		where (TIMESTAMPTZ 'epoch' + valid_to * '1 second'::interval) 
		between now() and now() + interval '30 days'
		and organization_id=$1
		and scan_group_id=$2
		and response_timestamp>=$3`,

	"webHashChanged": `select 
		wf.response_timestamp, 
		wf.url, 
		wf.host_address, 
		wf.response_port, 
		wf.prev_hash, 
		wf.serialized_dom_hash from 
			(select 
				response_timestamp, 
				url,
				host_address,
				response_port,
				lead(serialized_dom_hash) over (partition by url,response_port order by response_timestamp desc ) as prev_hash,
				row_number() over (partition by url,response_port order by response_timestamp desc ) as row_number,
				serialized_dom_hash from am.web_snapshots where 
				organization_id=$1 and
				scan_group_id=$2
			order by host_address,response_port,response_timestamp desc) as wf
		where row_number<=1;`,
}

var (
	AddTempTableKey     = "event_add_temp"
	AddTempTableColumns = []string{"organization_id", "scan_group_id", "type_id", "event_timestamp", "event_data", "event_data_json"}
	AddTempTable        = `create temporary table event_add_temp (
		organization_id int not null,
		scan_group_id int not null,
		type_id int not null,
		event_timestamp timestamptz not null,
		event_data jsonb,
		event_data_json json
		) on commit drop;`
	AddTempToAdd = `insert into am.event_notifications as unr (
		organization_id,
		scan_group_id,
		type_id,
		event_timestamp,
		event_data,
		event_data_json
	)
	select 
		temp.organization_id,
		temp.scan_group_id,
		temp.type_id,
		temp.event_timestamp,
		temp.event_data,
		temp.event_data_json
	from event_add_temp as temp`

	SubscriptionsTempTableKey     = "event_subs_temp"
	SubscriptionsTempTableColumns = []string{"organization_id", "user_id", "type_id", "subscribed_since", "subscribed"}
	SubscriptionsTempTable        = `create temporary table event_subs_temp (
			organization_id int not null,
			user_id int not null,
			type_id int not null,
			subscribed_since timestamptz not null,
			subscribed boolean not null
		) on commit drop;`

	SubscriptionsTempToSubscriptions = `insert into am.user_notification_subscriptions as unr (
		organization_id,
		user_id,
		type_id,
		subscribed_since,
		subscribed
	)
	select 
		temp.organization_id,
		temp.user_id,
		temp.type_id,
		temp.subscribed_since,
		temp.subscribed
	from event_subs_temp as temp on conflict (organization_id, user_id, type_id) do update set
		subscribed_since=EXCLUDED.subscribed_since,
		subscribed=EXCLUDED.subscribed`

	MarkReadTempTableKey     = "event_read_temp"
	MarkReadTempTableColumns = []string{"organization_id", "user_id", "notification_id"}
	MarkReadTempTable        = `create temporary table event_read_temp (
			organization_id int not null,
			user_id int not null,
			notification_id bigint not null
		) on commit drop;`

	MarkReadTempToMarkRead = `insert into am.user_notifications_read as unr (
		organization_id,
		user_id,
		notification_id
	)
	select 
		temp.organization_id,
		temp.user_id,
		temp.notification_id 
	from event_read_temp as temp on conflict do nothing`
)
