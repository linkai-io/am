package job

var queryMap = map[string]string{
	"getLastJob": `select job_id, organization_id, scan_group_id, job_timestamp, job_status from am.jobs 
		where organization_id=$1 and scan_group_id=$2 order by job_id desc limit 1`,

	"startJob": `insert into am.jobs (organization_id, scan_group_id, job_timestamp, job_status) 
		values ($1, $2, $3, $4) returning organization_id, job_id;`,

	"createJobEvent": `insert into am.job_events (organization_id, job_id, event_user_id, event_time, event_description, event_from)
		values ($1, $2, $3, $4, $5, $6)`,
}
