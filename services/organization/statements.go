package organization

import "fmt"

const (
	defaultColumns = `organization_name, organization_custom_id, user_pool_id, user_pool_client_id, 
	user_pool_client_secret, identity_pool_id, user_pool_jwk, owner_email, first_name, last_name, phone, country,
	state_prefecture, street, address1, address2, city, postal_code, creation_time,	status_id, deleted, subscription_id, 
	limit_tld, limit_tld_reached, limit_hosts, limit_hosts_reached, limit_custom_web_flows, limit_custom_web_flows_reached`
)

var queryMap = map[string]string{
	"orgExists": `select organization_id,organization_name,organization_custom_id from am.organizations where organization_name=$1 or organization_id=$2 or organization_custom_id=$3`,

	"orgByID": fmt.Sprintf(`select organization_id, %s from am.organizations where organization_id=$1`, defaultColumns),

	"orgByName": fmt.Sprintf(`select organization_id, %s from am.organizations where organization_name=$1`, defaultColumns),

	"orgByCID": fmt.Sprintf(`select organization_id, %s from am.organizations where organization_custom_id=$1`, defaultColumns),

	"orgByAppClientID": fmt.Sprintf(`select organization_id, %s from am.organizations where user_pool_client_id=$1`, defaultColumns),

	"orgList": fmt.Sprintf(`select organization_id, %s from am.organizations where organization_id > $1 order by organization_id limit $2`, defaultColumns),

	"orgCreate": fmt.Sprintf(`with org as (
					insert into am.organizations (%s
					)
					values 
						($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, false, $21, 
						$22, $23, $24, $25, $26, $27)
					returning organization_id
				) 
				insert into am.users (
						organization_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
					) 
					values
						( (select org.organization_id from org), $28, $29, $30, $31, $32, $33, false) returning organization_id,user_id;`, defaultColumns),

	// note this will call owner_user trigger to update am.users to keep in sync if email/first/last name changes.
	"orgUpdate": `update am.organizations set user_pool_id=$1, user_pool_client_id=$2, user_pool_client_secret=$3, 
			identity_pool_id=$4, user_pool_jwk=$5, owner_email=$6, first_name=$7, last_name=$8, phone=$9, country=$10, state_prefecture=$11, 
			street=$12, address1=$13, address2=$14, city=$15, postal_code=$16, status_id=$17, subscription_id=$18, 
			limit_tld=$19, limit_tld_reached=$20, limit_hosts=$21, limit_hosts_reached=$22, limit_custom_web_flows=$23, limit_custom_web_flows_reached=$24
			where organization_id=$25`,

	"orgDelete":      `update am.organizations set deleted=true, status_id=1 where organization_id=$1 returning organization_id`,
	"orgForceDelete": `select am.delete_org($1);`,
}
