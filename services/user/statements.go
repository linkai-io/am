package user

var queryMap = map[string]string{
	"userExists": `select organization_id, user_id, user_custom_id from am.users where organization_id=$1 and user_id=$2 or user_custom_id=$3`,

	"userByEmail": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
	from am.users where organization_id=$1 and email=$2`,

	"userByID": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
	from am.users where organization_id=$1 and user_id=$2`,

	"userByCID": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted 
	from am.users where organization_id=$1 and user_custom_id=$2`,

	"userList": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
	from am.users where organization_id=$1 and user_id > $2 order by user_id limit $3`,

	"userListWithDelete": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
	from am.users where organization_id=$1 and deleted=$2 and user_id > $3 order by user_id limit $4`,

	"userCreate": `insert into am.users (
						organization_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
					) 
					values
						($1, $2, $3, $4, $5, $6, $7, false) returning organization_id, user_id, user_custom_id;`,

	"userUpdate": `update am.users set email=$1, first_name=$2, last_name=$3, user_status_id=$4 where organization_id=$5 and user_id=$6 returning organization_id, user_id`,
	"userDelete": `update am.users set deleted=true, user_status_id=1 where organization_id=$1 and user_id=$2 returning organization_id`,
}
