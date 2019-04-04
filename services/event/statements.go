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
	"getUserSubscriptions": `select organization_id, user_id, type_id, subscribed_since, subscribed from am.user_notification_subscriptions where organization_id=$1 and user_id=$2`,
	"updateUserSettings": fmt.Sprintf(`insert into am.user_notification_settings (%s) values
		($1, $2, $3, $4, $5, $6, $7) on conflict (organization_id,user_id) do update set
		weekly_report_send_day=EXCLUDED.weekly_report_send_day,
		daily_report_send_hour=EXCLUDED.daily_report_send_hour,
		user_timezone=EXCLUDED.user_timezone,
		should_weekly_email=EXCLUDED.should_weekly_email,
		should_daily_email=EXCLUDED.should_daily_email`, sharedSettingColumns),

	"newHostnames": `select sga.host_address from 
		(select min(discovered_timestamp) as discovered_timestamp, host_address from am.scan_group_addresses 
			where organization_id=$1 and scan_group_id=$2 and 
			(confidence_score=100 or user_confidence_score=100) and
			deleted=false and
			ignored=false
			group by host_address) as uniq
		join am.scan_group_addresses as sga on sga.discovered_timestamp=uniq.discovered_timestamp 
			where sga.discovered_timestamp >= $3 and 
			organization_id=$4 and 
			scan_group_id=$5 group by sga.host_address`,
	"newWebsites": `select ws.url,ws.response_port from
		(select min(response_timestamp) as response_timestamp, url, response_port from am.web_snapshots 
			where organization_id=$1 and scan_group_id=$2 and 
			deleted=false
			group by url,response_port) as uniq
		join am.web_snapshots as ws on ws.response_timestamp=uniq.response_timestamp
			where ws.response_timestamp >= $3 and
			organization_id=$4 and
			scan_group_id=$5 group by ws.url, ws.response_port`,
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
	"checkCertExpiration": `select subject_name, port, valid_to from am.web_certificates 
		where (TIMESTAMPTZ 'epoch' + valid_to * '1 second'::interval) 
		between now() and now() + interval '30 days'
		and organization_id=$1
		and scan_group_id=$2
		and response_timestamp>=$3`,
}

var (
	AddTempTableKey     = "event_add_temp"
	AddTempTableColumns = []string{"organization_id", "scan_group_id", "type_id", "event_timestamp", "event_data"}
	AddTempTable        = `create temporary table event_add_temp (
		organization_id int not null,
		scan_group_id int not null,
		type_id int not null,
		event_timestamp timestamptz not null,
		event_data jsonb
		) on commit drop;`
	AddTempToAdd = `insert into am.event_notifications as unr (
		organization_id,
		scan_group_id,
		type_id,
		event_timestamp,
		event_data
	)
	select 
		temp.organization_id,
		temp.scan_group_id,
		temp.type_id,
		temp.event_timestamp,
		temp.event_data 
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
