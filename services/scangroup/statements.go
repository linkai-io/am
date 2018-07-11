package scangroup

// TODO: make a struct instead of map
var queryMap = map[string]string{
	// am.scan_group related
	"scanGroupIDByName": "select organization_id, scan_group_id from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false",
	"scanGroupByName":   "select organization_id, scan_group_id, scan_group_name, creation_time, created_by, original_input, deleted from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false",
	"scanGroupByID":     "select organization_id, scan_group_id, scan_group_name, creation_time, created_by, original_input, deleted from am.scan_group where organization_id=$1 and scan_group_id=$2 and deleted=false",
	"scanGroupsByOrgID": "select organization_id, scan_group_id, scan_group_name, creation_time, created_by, original_input, deleted from am.scan_group where organization_id=$1 and deleted=false",
	"deleteScanGroup":   "update am.scan_group set deleted=true where organization_id=$1 and scan_group_id=$2",
	"createScanGroup": `insert into am.scan_group (organization_id, scan_group_name, creation_time, created_by, original_input, deleted) values 
		($1, $2, $3, $4, $5, false) returning scan_group_id`,

	// am.scan_group_versions related
	"scanGroupVersionExists": "select organization_id, scan_group_version_id from am.scan_group_versions where organization_id=$1 and scan_group_id=$2 and deleted=false",
	"scanGroupVersionIDs":    "select organization_id, scan_group_id, scan_group_version_id from am.scan_group_versions where organization_id=$1 and scan_group_id=$2 and deleted=false",
	"scanGroupVersion":       "select organization_id, scan_group_id, scan_group_version_id, version_name, creation_time, created_by, configuration, deleted from am.scan_group_versions where organization_id=$1 and scan_group_id=$2 and deleted=false",
	"scanGroupVersionByID":   "select organization_id, scan_group_id, scan_group_version_id, version_name, creation_time, created_by, configuration, deleted from am.scan_group_versions where organization_id=$1 and scan_group_id=$2 and scan_group_version_id=$3 and deleted=false",
	"scanGroupVersionByName": "select organization_id, scan_group_id, scan_group_version_id, version_name, creation_time, created_by, configuration, deleted from am.scan_group_versions where organization_id=$1 and scan_group_id=$2 and version_name=$3 and deleted=false",
	"deleteScanGroupVersion": "update am.scan_group_versions set deleted=true where organization_id=$1 and scan_group_id=$2 and scan_group_version_id=$3",
	"createScanGroupVersion": `insert into am.scan_group_versions (organization_id, scan_group_id, version_name, creation_time, created_by, configuration, deleted) values
		($1, $2, $3, $4, $5, $6, false) returning scan_group_version_id`,

	// am.scan_group_addresses related
	"scanGroupAddresses":              "select * from am.scan_group_addresses where organization_id=$1 and scan_group_id=$2",
	"scanGroupAddressesFilterIgnored": "select * from am.scan_group_addresses where organization_id=$1 and scan_group_id=$2 and is_ignored=false",
	//"scanGroupAddAddresses":           "",
	//"scanGroupUpdateAddresses":        "",
}
