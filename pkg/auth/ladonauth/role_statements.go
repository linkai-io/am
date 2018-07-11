package ladonauth

// RoleStatements hold queries necessary for the LadonRoleManager
type RoleStatements struct {
	QueryAddMembers    string
	QueryDeleteMembers string
	QueryDeleteRole    string
	QueryFindByMember  string
	QueryGetMember     string
	QueryGetRole       string
	QueryGetRoleByName string
	QueryInsertRole    string
	QueryList          string
}

// GetRoleStatements populates the statements structure with our queries specific to roles
func GetRoleStatements(driverName string) *RoleStatements {
	stmts := &RoleStatements{}
	switch driverName {
	case "postgres", "pgx", "pg":
		stmts.QueryAddMembers = "INSERT INTO am.ladon_role_member (organization_id, role_id, member_id) VALUES ($1, $2, $3)"
		stmts.QueryDeleteMembers = "DELETE FROM am.ladon_role_member WHERE organization_id=$1 and member_id=$2 AND role_id=$3"
		stmts.QueryDeleteRole = "DELETE FROM am.ladon_role WHERE organization_id=$1 and role_id=$2"
		stmts.QueryFindByMember = "SELECT organization_id,role_id from am.ladon_role_member WHERE organization_id=$1 and member_id=$2 GROUP BY organization_id,role_id ORDER BY role_id LIMIT $3 OFFSET $4"
		stmts.QueryGetMember = "SELECT organization_id,member_id from am.ladon_role_member WHERE organization_id=$1 and role_id=$2"
		stmts.QueryGetRole = "SELECT organization_id,role_id,role_name from am.ladon_role WHERE organization_id=$1 and role_id=$2"
		stmts.QueryGetRoleByName = "SELECT organization_id,role_id,role_name from am.ladon_role WHERE organization_id=$1 and role_name=$2"
		stmts.QueryInsertRole = "INSERT INTO am.ladon_role (organization_id, role_id, role_name) VALUES ($1,$2,$3)"
		stmts.QueryList = "SELECT organization_id,role_id,role_name from am.ladon_role WHERE organization_id=$1 GROUP BY role_id ORDER BY role_id LIMIT $2 OFFSET $3"
	default:
		return nil
	}
	return stmts
}
