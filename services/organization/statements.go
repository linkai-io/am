package organization

var queryMap = map[string]string{
	"orgExists": `select organization_id,organization_name,organization_custom_id from am.organizations where organization_name=$1 or organization_id=$2 or organization_custom_id=$3`,

	"orgByID": `select 
		organization_id, organization_name, organization_custom_id, user_pool_id,
		identity_pool_id, owner_email, first_name, last_name, phone, country,
		state_prefecture, street, address1, address2, city, postal_code, creation_time,
		status_id, deleted, subscription_id
	from am.organizations where organization_id=$1`,

	"orgByName": `select 
		organization_id, organization_name, organization_custom_id, user_pool_id,
		identity_pool_id, owner_email, first_name, last_name, phone, country,
		state_prefecture, street, address1, address2, city, postal_code, creation_time,
		status_id, deleted, subscription_id
	from am.organizations where organization_name=$1`,

	"orgByCID": `select 
		organization_id, organization_name, organization_custom_id, user_pool_id,
		identity_pool_id, owner_email, first_name, last_name, phone, country,
		state_prefecture, street, address1, address2, city, postal_code, creation_time,
		status_id, deleted, subscription_id
	from am.organizations where organization_custom_id=$1`,

	"orgList": `select 
		organization_id, organization_name, organization_custom_id, user_pool_id,
		identity_pool_id, owner_email, first_name, last_name, phone, country,
		state_prefecture, street, address1, address2, city, postal_code, creation_time,
		status_id, deleted, subscription_id
	from am.organizations where organization_id > $1 order by organization_id limit $2`,

	"orgCreate": `with org as (
					insert into am.organizations (
						organization_name, organization_custom_id, user_pool_id, identity_pool_id, 
						owner_email, first_name, last_name, phone, country, state_prefecture, street, 
						address1, address2, city, postal_code, creation_time, status_id, deleted, subscription_id
					)
					values 
						($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, false, $18)
					returning organization_id
				) 
				insert into am.users (
						organization_id, user_custom_id, email, first_name, last_name, deleted
					) 
					values
						( (select org.organization_id from org), $19, $20, $21, $22, false);`,

	"orgUpdate": `update am.organizations set deleted=$1 where organization_id=$2`,
	"orgDelete": `update am.organizations set deleted=true, status_id=1 where organization_id=$1`,
}
