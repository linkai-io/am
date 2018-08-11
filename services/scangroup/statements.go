package scangroup

var queryMap = map[string]string{
	// am.scan_group related
	"scanGroupByID": `select organization_id, scan_group_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration, paused, deleted
	 	from am.scan_group where organization_id=$1 and scan_group_id=$2 and deleted=false`,

	"scanGroupIDByName": "select organization_id, scan_group_id from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false",

	"scanGroupByName": `select organization_id, scan_group_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration, paused, deleted 
	 	from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false`,

	"scanGroupsByOrgID": `select organization_id, scan_group_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration, paused, deleted  
		from am.scan_group where organization_id=$1 and deleted=false`,

	"scanGroupName": `select organization_id, scan_group_name from am.scan_group where organization_id=$1 and scan_group_id=$2`,

	// updates the scan_group_name to name_<deleted_timestamp>
	"deleteScanGroup": "update am.scan_group set deleted=true, scan_group_name=$1 where organization_id=$2 and scan_group_id=$3",

	"createScanGroup": `insert into am.scan_group (organization_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration, paused, deleted) values 
		($1, $2, $3, $4, $5, $6, $7, $8, false, false) returning organization_id, scan_group_id`,

	"updateScanGroup": `update am.scan_group set scan_group_name=$1, modified_time=$2, modified_by=$3, configuration=$4 
		where organization_id=$5 and scan_group_id=$6 returning organization_id, scan_group_id`,

	"pauseScanGroup": `update am.scan_group set paused=true, modified_time=$1, modified_by=$2 
		where organization_id=$3 and scan_group_id=$4 returning organization_id, scan_group_id`,

	"resumeScanGroup": `update am.scan_group set paused=false, modified_time=$1, modified_by=$2 
		where organization_id=$3 and scan_group_id=$4 returning organization_id, scan_group_id`,
}
