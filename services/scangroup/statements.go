package scangroup

import "fmt"

const (
	defaultColumns = `organization_id, scan_group_id, scan_group_name, 
	creation_time, (select email from am.users where am.users.user_id=created_by) as created_by_user, created_by,
	modified_time, (select email from am.users where am.users.user_id=modified_by) as modified_by_user, modified_by,
	original_input_s3_url, configuration, paused, deleted, last_paused_timestamp, archive_after_days`
)

var queryMap = map[string]string{
	// am.scan_group related
	"allScanGroups": fmt.Sprintf(`select %s from am.scan_group where deleted=false`, defaultColumns),

	"allScanGroupsWithPaused": fmt.Sprintf(`select %s from am.scan_group where deleted=false and paused=$1`, defaultColumns),

	"scanGroupByID": fmt.Sprintf(`select %s	from am.scan_group where organization_id=$1 and scan_group_id=$2 and deleted=false`, defaultColumns),

	"scanGroupByName": fmt.Sprintf(`select %s from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false`, defaultColumns),

	"scanGroupsByOrgID": fmt.Sprintf(`select %s from am.scan_group where organization_id=$1 and deleted=false`, defaultColumns),

	"scanGroupIDByName": "select organization_id, scan_group_id from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false",

	"scanGroupName": `select organization_id, scan_group_name from am.scan_group where organization_id=$1 and scan_group_id=$2`,

	// updates the scan_group_name to name_<deleted_timestamp>
	"deleteScanGroup": "update am.scan_group set deleted=true, scan_group_name=$1 where organization_id=$2 and scan_group_id=$3",

	"createScanGroup": `insert into am.scan_group (organization_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input_s3_url, configuration, paused, deleted, archive_after_days) values 
		($1, $2, $3, $4, $5, $6, $7, $8, false, false, $9) returning organization_id, scan_group_id`,

	"updateScanGroup": `update am.scan_group set scan_group_name=$1, modified_time=$2, modified_by=$3, configuration=$4, archive_after_days=$5 
		where organization_id=$6 and scan_group_id=$7 returning organization_id, scan_group_id`,

	"pauseScanGroup": `update am.scan_group set paused=true, last_paused_timestamp=now(), modified_time=$1, modified_by=$2 
		where organization_id=$3 and scan_group_id=$4 returning organization_id, scan_group_id`,

	"resumeScanGroup": `update am.scan_group set paused=false, modified_time=$1, modified_by=$2 
		where organization_id=$3 and scan_group_id=$4 returning organization_id, scan_group_id`,

	"updateGroupActivity": `insert into am.scan_group_activity (organization_id, scan_group_id, active_addresses, batch_size, last_updated, batch_start, batch_end) values 
		($1, $2, $3, $4, $5, $6, $7) on conflict (organization_id, scan_group_id) do update set
		active_addresses=EXCLUDED.active_addresses,
		batch_size=EXCLUDED.batch_size,
		last_updated=EXCLUDED.last_updated,
		batch_start=EXCLUDED.batch_start,
		batch_end=EXCLUDED.batch_end`,

	"getOrgGroupActivity": `select organization_id, scan_group_id, active_addresses, batch_size, last_updated, batch_start, batch_end from am.scan_group_activity 
		where organization_id=$1 and scan_group_id in (select scan_group_id from am.scan_group where organization_id=$1 and deleted=false)`,
}
