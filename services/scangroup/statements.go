package scangroup

import "fmt"

const (
	defaultColumns = `organization_id, scan_group_id, scan_group_name, 
	creation_time, (select email from am.users where am.users.user_id=created_by) as created_by_user, created_by,
	modified_time, (select email from am.users where am.users.user_id=modified_by) as modified_by_user, modified_by,
	original_input_s3_url, configuration, paused, deleted`
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

	"createScanGroup": `insert into am.scan_group (organization_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input_s3_url, configuration, paused, deleted) values 
		($1, $2, $3, (select am.users.user_id from am.users where email=$4 and am.users.organization_id=$1), $5, (select am.users.user_id from am.users where email=$6 and am.users.organization_id=$1), $7, $8, false, false) returning organization_id, scan_group_id`,

	"updateScanGroup": `update am.scan_group set scan_group_name=$1, modified_time=$2, modified_by=(select am.users.user_id from am.users where email=$3 and am.users.organization_id=$5), configuration=$4 
		where organization_id=$5 and scan_group_id=$6 returning organization_id, scan_group_id`,

	"pauseScanGroup": `update am.scan_group set paused=true, modified_time=$1, modified_by=$2 
		where organization_id=$3 and scan_group_id=$4 returning organization_id, scan_group_id`,

	"resumeScanGroup": `update am.scan_group set paused=false, modified_time=$1, modified_by=$2 
		where organization_id=$3 and scan_group_id=$4 returning organization_id, scan_group_id`,
}
