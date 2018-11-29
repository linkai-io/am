package organization

import "fmt"

const (
	defaultColumns = `organization_name, organization_custom_id, user_pool_id, user_pool_client_id, 
	user_pool_client_secret, identity_pool_id, owner_email, first_name, last_name, phone, country,
	state_prefecture, street, address1, address2, city, postal_code, creation_time,	status_id, deleted, subscription_id`
)

var queryMap = map[string]string{
	"orgExists": `select organization_id,organization_name,organization_custom_id from am.organizations where organization_name=$1 or organization_id=$2 or organization_custom_id=$3`,

	"orgByID": fmt.Sprintf(`select organization_id, %s from am.organizations where organization_id=$1`, defaultColumns),

	"orgByName": fmt.Sprintf(`select organization_id, %s from am.organizations where organization_name=$1`, defaultColumns),

	"orgByCID": fmt.Sprintf(`select organization_id, %s from am.organizations where organization_custom_id=$1`, defaultColumns),

	"orgList": fmt.Sprintf(`select organization_id, %s from am.organizations where organization_id > $1 order by organization_id limit $2`, defaultColumns),

	"orgCreate": fmt.Sprintf(`with org as (
					insert into am.organizations (%s
					)
					values 
						($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, false, $20)
					returning organization_id
				) 
				insert into am.users (
						organization_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
					) 
					values
						( (select org.organization_id from org), $21, $22, $23, $24, $25, $26, false) returning organization_id,user_id;`, defaultColumns),

	// note this will call owner_user trigger to update am.users to keep in sync if email/first/last name changes.
	"orgUpdate": `update am.organizations set user_pool_id=$1, user_pool_client_id=$2, user_pool_client_secret=$3, 
			identity_pool_id=$4, owner_email=$5, first_name=$6,	last_name=$7, phone=$8, country=$9, state_prefecture=$10, 
			street=$11, address1=$12, address2=$13, city=$14, postal_code=$15, status_id=$16, subscription_id=$17 where organization_id=$18`,

	"orgDelete":      `update am.organizations set deleted=true, status_id=1 where organization_id=$1 returning organization_id`,
	"orgForceDelete": `select am.delete_org($1);`,
}
