package pg

var queryMap = map[string]string{
	// user related
	"userRole": "select role_id from am.users where organization_id=$1 and user_id=$2",

	// am.scan_group related
	"scanGroupIDByName": "select scan_group_id from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false",
	"scanGroupByName":   "select * from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false",
	"scanGroupByID":     "select * from am.scan_group where organization_id=$1 and scan_group_id=$2 and deleted=false",
	"scanGroupsByOrgID": "select * from am.scan_group where organization_id=$1 and deleted=false",
	"deleteScanGroup":   "update am.scan_group set deleted=true where organization_id=$1 and scan_group_id=$2",
	"createScanGroup": `insert into am.scan_group (organization_id, scan_group_name, creation_time, created_by, original_input, deleted) values 
		($1, $2, $3, $4, $5, false)`,

	// am.scan_group_versions related
	"scanGroupVersion":       "select * from am.scan_group_versions where organization_id=$1 and scan_group_id=$2 and deleted=false",
	"scanGroupVersionByID":   "select * from am.scan_group_versions where organization_id=$1 and scan_group_id=$2 and scan_group_version_id=$3 and deleted=false",
	"scanGroupVersionByName": "select * from am.scan_group_versions where organization_id=$1 and scan_group_id=$2 and version_name=$3 and deleted=false",
	"deleteScanGroupVersion": "update am.scan_group_versions set deleted=true where organization_id=$1 and scan_group_id=$2 and scan_group_version_id=$3",
	"createScanGroupVersion": `insert into am.scan_group_versions (organization_id, scan_group_id, version_name, creation_time, created_by, configuration, config_version, deleted) values
		($1, $2, $3, $4, $5, $6, $7, false)`,

	// am.scan_group_addresses related
	"scanGroupAddresses":              "select * from am.scan_group_addresses where organization_id=$1 and scan_group_id=$2",
	"scanGroupAddressesFilterIgnored": "select * from am.scan_group_addresses where organization_id=$1 and scan_group_id=$2 and is_ignored=false",
}
