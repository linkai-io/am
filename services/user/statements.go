package user

var queryMap = map[string]string{
	"userExists": `select organization_id,user_id,user_custom_id from am.users where organization_id=$1 and user_id=$2 or user_custom_id=$3`,

	"userByID": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, deleted
	from am.users where organization_id=$1 and user_id=$2`,

	"userByCID": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, deleted
	from am.users where organization_id=$1 and user_custom_id=$2`,

	"userList": `select 
		organization_id, user_id, user_custom_id, email, first_name, last_name, deleted
	from am.users where organization_id=$1 and user_id > $1 order by user_id limit $2`,

	"userCreate": `insert into am.users (
						organization_id, user_custom_id, email, first_name, last_name, deleted
					) 
					values
						($1, $2, $3, $4, $5, false);`,

	"userUpdate": `update am.users set deleted=$1 where organization_id=$2`,
	"userDelete": `update am.users set deleted=true where organization_id=$1 and user_id=$2`,
}
