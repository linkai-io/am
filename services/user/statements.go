package user

var queryMap = map[string]string{
	"userExists": `select organization_id, user_id, user_custom_id from am.users where organization_id=$1 and user_id=$2 or user_custom_id=$3`,

	"userByEmail": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted, agreement_accepted, agreement_accepted_timestamp
	from am.users where organization_id=$1 and email=$2`,

	"userByID": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted, agreement_accepted, agreement_accepted_timestamp
	from am.users where organization_id=$1 and user_id=$2`,

	"userByCID": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted, agreement_accepted, agreement_accepted_timestamp 
	from am.users where organization_id=$1 and user_custom_id=$2`,

	"userList": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted, agreement_accepted, agreement_accepted_timestamp
	from am.users where organization_id=$1 and user_id > $2 order by user_id limit $3`,

	"userListWithDelete": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted, agreement_accepted, agreement_accepted_timestamp
	from am.users where organization_id=$1 and deleted=$2 and user_id > $3 order by user_id limit $4`,

	"userCreate": `insert into am.users (
						organization_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
					) 
					values
						($1, $2, $3, $4, $5, $6, $7, false) returning organization_id, user_id, user_custom_id;`,

	"userUpdate":          `update am.users set user_custom_id=$1, email=$2, first_name=$3, last_name=$4, user_status_id=$5 where organization_id=$6 and user_id=$7 returning organization_id, user_id`,
	"userUpdateAgreement": `update am.users set agreement_accepted=true, agreement_accepted_timestamp=now() where organization_id=$1 and user_id=$2 and agreement_accepted=false`,
	"userDelete":          `update am.users set deleted=true, user_status_id=1 where organization_id=$1 and user_id=$2 returning organization_id`,
}
