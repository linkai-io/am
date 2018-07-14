package scangroup

var queryMap = map[string]string{
	// am.scan_group related
	"scanGroupByID": `select organization_id, scan_group_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration, deleted
	 	from am.scan_group where organization_id=$1 and scan_group_id=$2 and deleted=false`,

	"scanGroupIDByName": "select organization_id, scan_group_id from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false",

	"scanGroupByName": `select organization_id, scan_group_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration, deleted
	 	from am.scan_group where organization_id=$1 and scan_group_name=$2 and deleted=false`,

	"scanGroupsByOrgID": `select organization_id, scan_group_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration, deleted 
		from am.scan_group where organization_id=$1 and deleted=false`,

	"scanGroupName": `select organization_id, scan_group_name from am.scan_group where organization_id=$1 and scan_group_id=$2`,

	// updates the scan_group_name to name_<deleted_timestamp>
	"deleteScanGroup": "update am.scan_group set deleted=true, scan_group_name=$1 where organization_id=$2 and scan_group_id=$3",

	"createScanGroup": `insert into am.scan_group (organization_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration, deleted) values 
		($1, $2, $3, $4, $5, $6, $7, $8, false) returning organization_id, scan_group_id`,

	"updateScanGroup": `update am.scan_group set scan_group_name=$1, modified_time=$2, modified_by=$3, configuration=$4 
		where organization_id=$5 and scan_group_id=$6 returning organization_id, scan_group_id`,

	// am.scan_group_addresses related
	"scanGroupAddressesCount": `select count(address_id) as count from am.scan_group_addresses where organization_id=$1 and scan_group_id=$2 and deleted=false`,

	// returns
	"scanGroupAddressesFilterIgnoredDeleted": `select 
		organization_id, 
		address_id, 
		scan_group_id, 
		address, 
		added_timestamp, 
		(select added_by from am.scan_address_added_by where scan_address_added_id=sga.scan_address_added_id),
		ignored,
		deleted
		from am.scan_group_addresses as sga where organization_id=$1 and scan_group_id=$2 and ignored=false and deleted=true and address_id > $5 order by address_id limit $6`,

	"scanGroupAddresses": `select 
		organization_id, 
		address_id, 
		scan_group_id, 
		address, 
		added_timestamp, 
		(select added_by from am.scan_address_added_by where scan_address_added_id=sga.scan_address_added_id),
		ignored,
		deleted
		from am.scan_group_addresses as sga where organization_id=$1 and scan_group_id=$2 and deleted=false and address_id > $3 order by address_id limit $4`,

	"scanGroupAddressesIgnored": `select 
		organization_id, 
		address_id, 
		scan_group_id, 
		address, 
		added_timestamp, 
		(select added_by from am.scan_address_added_by where scan_address_added_id=sga.scan_address_added_id),
		ignored,
		deleted
		from am.scan_group_addresses as sga where organization_id=$1 and scan_group_id=$2 and ignored=true and address_id > $3 order by address_id limit $4`,

	"scanGroupAddressesDeleted": `select 
		organization_id, 
		address_id, 
		scan_group_id, 
		address, 
		added_timestamp, 
		(select added_by from am.scan_address_added_by where scan_address_added_id=sga.scan_address_added_id),
		ignored,
		deleted
		from am.scan_group_addresses as sga where organization_id=$1 and scan_group_id=$2 and deleted=true and address_id > $3 order by address_id limit $4`,
}

var (
	AddAddressesTempTableKey     = "sga_add_temp"
	AddAddressesTempTableColumns = []string{"organization_id", "scan_group_id", "address", "added_timestamp", "scan_address_added_by", "deleted", "ignored"}
	AddAddressesTempTable        = `create temporary table sga_add_temp (
		organization_id integer not null,
		scan_group_id integer not null,
		address varchar(512) not null,
		added_timestamp bigint not null,
		scan_address_added_by varchar(128) not null,
		deleted boolean,
		ignored boolean
		) on commit drop;`

	AddAddressesTempToAddress = `insert into am.scan_group_addresses as sga (
			organization_id, 
			scan_group_id,
			address,
			added_timestamp,
			scan_address_added_id,
			deleted,
			ignored
		)
		select
			st.organization_id, 
			st.scan_group_id, 
			st.address, 
			st.added_timestamp, 
			(select scan_address_added_id from am.scan_address_added_by where added_by=st.scan_address_added_by),
			st.deleted,
			st.ignored 
		from sga_add_temp as st on conflict do nothing;`

	// for ignoring/unignoring addresses
	IgnoreAddressesTempTableKey     = "sga_ignored_temp"
	IgnoreAddressesTempTableColumns = []string{"organization_id", "scan_group_id", "address_id", "ignored"}
	IgnoreAddressesTempTable        = `create temporary table sga_ignored_temp (
		organization_id integer not null,
		scan_group_id integer not null,
		address_id bigint,
		ignored boolean
		) on commit drop;`

	/*
				UPDATE tableone
		SET field1 = tabletwo.fieldx
		FROM tabletwo
		WHERE tableone.commonid = tabletwo.commonid;
	*/
	IgnoreAddressesTempToAddress = `update am.scan_group_addresses as sga
		set ignored=sga_ignored_temp.ignored 
		from sga_ignored_temp where sga.address_id=sga_ignored_temp.address_id;`

	// for deleting addresses
	DeleteAddressesTempTableKey     = "sga_deleted_temp"
	DeleteAddressesTempTableColumns = []string{"organization_id", "scan_group_id", "address_id", "deleted"}
	DeleteAddressesTempTable        = `create temporary table sga_deleted_temp (
		organization_id integer not null,
		scan_group_id integer not null,
		address_id bigint,
		deleted boolean
		) on commit drop;`

	DeleteAddressesTempToAddress = `update am.scan_group_addresses as sga
		set deleted=sga_deleted_temp.deleted, address=sga.address||$1 
		from sga_deleted_temp where sga.address_id=sga_deleted_temp.address_id;`
)
