package job

var queryMap = map[string]string{
	"userExists": `select organization_id, user_id, user_custom_id from am.users where organization_id=$1 and user_id=$2 or user_custom_id=$3`,
}
