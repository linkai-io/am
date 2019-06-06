package user

import "fmt"

const commonColumns = `organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted, agreement_accepted, agreement_accepted_timestamp, last_login_timestamp`

var queryMap = map[string]string{
	"userExists": `select organization_id, user_id, user_custom_id from am.users where organization_id=$1 and user_id=$2 or user_custom_id=$3`,

	"userByEmail": fmt.Sprintf(`select %s from am.users where organization_id=$1 and email=$2`, commonColumns),

	"userByID": fmt.Sprintf(`select %s from am.users where organization_id=$1 and user_id=$2`, commonColumns),

	"userByCID": fmt.Sprintf(`select %s	from am.users where organization_id=$1 and user_custom_id=$2`, commonColumns),

	"userList": fmt.Sprintf(`select %s from am.users where organization_id=$1 and user_id > $2 order by user_id limit $3`, commonColumns),

	"userListWithDelete": fmt.Sprintf(`select %s from am.users where organization_id=$1 and deleted=$2 and user_id > $3 order by user_id limit $4`, commonColumns),

	"userCreate": `insert into am.users (
						organization_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted
					) 
					values
						($1, $2, $3, $4, $5, $6, $7, false) returning organization_id, user_id, user_custom_id;`,

	"userUpdate":          `update am.users set user_custom_id=$1, email=$2, first_name=$3, last_name=$4, user_status_id=$5, last_login_timestamp=$6 where organization_id=$7 and user_id=$8 returning organization_id, user_id`,
	"userUpdateAgreement": `update am.users set agreement_accepted=true, agreement_accepted_timestamp=now() where organization_id=$1 and user_id=$2 and agreement_accepted=false`,
	"userDelete":          `update am.users set deleted=true, user_status_id=1 where organization_id=$1 and user_id=$2 returning organization_id`,
}
