package ladonrolemanager

type Statements struct {
	QueryAddMembers    string
	QueryDeleteMembers string
	QueryDeleteRole    string
	QueryFindByMember  string
	QueryGetMember     string
	QueryGetRole       string
	QueryInsertRole    string
	QueryList          string
}

func GetStatements(driverName string) *Statements {
	stmts := &Statements{}
	switch driverName {
	case "postgres", "pgx", "pg":
		stmts.QueryAddMembers = "INSERT INTO am.ladon_role_member (organization_id, role_id, member) VALUES ($1, $2, $3)"
		stmts.QueryDeleteMembers = "DELETE FROM am.ladon_role_member WHERE organization_id=$1 and member=$2 AND role_id=$3"
		stmts.QueryDeleteRole = "DELETE FROM am.ladon_role WHERE organization_id=$1 and role_id=$2"
		stmts.QueryFindByMember = "SELECT organization_id,role_id from am.ladon_role_member WHERE organization_id=$1 and member=$2 GROUP BY role_id ORDER BY role_id LIMIT $3 OFFSET $4"
		stmts.QueryGetMember = "SELECT organization_id,member from am.ladon_role_member WHERE organization_id=$1 and role_id=$2"
		stmts.QueryGetRole = "SELECT organization_id,role_id from am.ladon_role WHERE organization_id=$1 and role_id=$2"
		stmts.QueryInsertRole = "INSERT INTO am.ladon_role (organization_id, role_id) VALUES ($1,$2)"
		stmts.QueryList = "SELECT organization_id,role_id from am.ladon_role WHERE organization_id=$1 GROUP BY role_id ORDER BY role_id LIMIT $2 OFFSET $3"
	default:
		return nil
	}
	return stmts
}
