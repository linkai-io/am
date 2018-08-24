package ladonauth

import (
	"context"
	"errors"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/linkai-io/am/am"
)

var (
	// ErrTxCreateFailed returned if we could not begin a transaction
	ErrTxCreateFailed = errors.New("could not begin transaction")
	// ErrRoleNotFound unable to find role for orgID
	ErrRoleNotFound = errors.New("role id does not exist")
	// ErrMismatchOrgID if a query ever returns data that doesn't match the requester orgID
	ErrMismatchOrgID = errors.New("org id returned does not match specified org id")
	// ErrMissingRoleName if a new role/updated role is missing the role name.
	ErrMissingRoleName = errors.New("missing role_name from role")
	// ErrMissingOrgID if a new role is missing the orgid.
	ErrMissingOrgID = errors.New("missing org id from new role")
	// ErrMemberNotFound if a member does not exist in any roles
	ErrMemberNotFound = errors.New("member was not found in any roles defined")
)

// LadonRoleManager manages user roles and groups
type LadonRoleManager struct {
	db         *pgx.ConnPool
	driverName string
	stmts      *RoleStatements
}

// NewRoleManager returns a new LadonRoleManager
func NewRoleManager(db *pgx.ConnPool, driverName string) *LadonRoleManager {
	return &LadonRoleManager{db: db, driverName: driverName}
}

// Init this role manager with a pgx connection pool
// TODO: test db tables exist
func (r *LadonRoleManager) Init() error {
	if r.stmts == nil {
		r.stmts = GetRoleStatements(r.driverName)
		if r.stmts == nil {
			return ErrInvalidDriver
		}
	}
	return nil
}

// CreateRole with or without initial members.
// returns roleID
func (r *LadonRoleManager) CreateRole(g *am.Role) (string, error) {
	if g.RoleName == "" {
		return "", ErrMissingRoleName
	}

	if g.OrgID == 0 {
		return "", ErrMissingOrgID
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	g.ID = id.String()

	if _, err = r.db.Exec(r.stmts.QueryInsertRole, g.OrgID, g.ID, g.RoleName); err != nil {
		return "", err
	}
	return g.ID, r.AddMembers(g.OrgID, g.ID, g.Members)
}

// DeleteRole from the system, will return non-error if orgID and roleID are invalid
func (r *LadonRoleManager) DeleteRole(orgID int, roleID string) error {
	_, err := r.db.Exec(r.stmts.QueryDeleteRole, orgID, roleID)
	return err
}

// AddMembers iterates over every member for the group/roleid and adds it to the ladon_role_members table
func (r *LadonRoleManager) AddMembers(orgID int, roleID string, members []int) error {
	if members == nil || len(members) == 0 {
		return nil
	}

	b := r.db.BeginBatch()

	for _, member := range members {
		b.Queue(r.stmts.QueryAddMembers, []interface{}{orgID, roleID, member}, []pgtype.OID{pgtype.Int4OID, pgtype.VarcharOID, pgtype.Int4OID}, nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := b.Send(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := b.ExecResults(); err != nil {
		return err
	}

	return nil
}

// RemoveMembers iterates over every member for the group/roleid and removes it from the ladon_role_members table
func (r *LadonRoleManager) RemoveMembers(orgID int, roleID string, members []int) error {
	if members == nil || len(members) == 0 {
		return nil
	}

	b := r.db.BeginBatch()

	for _, member := range members {
		b.Queue(r.stmts.QueryDeleteMembers, []interface{}{orgID, member, roleID}, []pgtype.OID{pgtype.Int4OID, pgtype.Int4OID, pgtype.VarcharOID}, nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := b.Send(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := b.ExecResults(); err != nil {
		return err
	}

	return nil
}

// FindByMember returns roles that a member belongs to
func (r *LadonRoleManager) FindByMember(orgID int, member int, limit, offset int) ([]*am.Role, error) {
	var roleIDs []string
	rows, err := r.db.Query(r.stmts.QueryFindByMember, orgID, member, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var rOrgID int

		if err := rows.Scan(&rOrgID, &id); err != nil {
			return nil, err
		}

		if rOrgID != orgID {
			return nil, ErrMismatchOrgID
		}

		roleIDs = append(roleIDs, id)
	}

	if len(roleIDs) == 0 {
		return nil, ErrMemberNotFound
	}

	var groups = make([]*am.Role, len(roleIDs))
	for k, roleID := range roleIDs {
		group, err := r.Get(orgID, roleID)
		if err != nil {
			return nil, err
		}

		groups[k] = group
	}

	return groups, nil
}

// Get a role specified by roleID for the orgID
func (r *LadonRoleManager) Get(orgID int, roleID string) (*am.Role, error) {
	return r.getBy(orgID, roleID, "")
}

// GetByName a role specified by roleName for the orgID
func (r *LadonRoleManager) GetByName(orgID int, roleName string) (*am.Role, error) {
	return r.getBy(orgID, "", roleName)
}

// List all roles for an organization
func (r *LadonRoleManager) List(orgID int, limit, offset int) ([]*am.Role, error) {
	var roleIDs []string
	rows, err := r.db.Query(r.stmts.QueryList, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var roleName string
		var rOrgID int

		if err := rows.Scan(&rOrgID, &id, &roleName); err != nil {
			return nil, err
		}

		if rOrgID != orgID {
			return nil, ErrMismatchOrgID
		}

		roleIDs = append(roleIDs, id)
	}

	var groups = make([]*am.Role, len(roleIDs))
	for k, roleID := range roleIDs {
		group, err := r.Get(orgID, roleID)
		if err != nil {
			return nil, err
		}

		groups[k] = group
	}

	return groups, nil
}

func (r *LadonRoleManager) getBy(orgID int, roleID, roleName string) (*am.Role, error) {
	var found string
	var foundName string
	var rOrgID int

	if roleID == "" && roleName == "" {
		return nil, ErrRoleNotFound
	}

	if roleName == "" {
		if err := r.db.QueryRow(r.stmts.QueryGetRole, orgID, roleID).Scan(&rOrgID, &found, &foundName); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.QueryRow(r.stmts.QueryGetRoleByName, orgID, roleName).Scan(&rOrgID, &found, &foundName); err != nil {
			return nil, err
		}
	}

	if found == "" {
		return nil, ErrRoleNotFound
	}

	if rOrgID != orgID {
		return nil, ErrMismatchOrgID
	}

	rows, err := r.db.Query(r.stmts.QueryGetMember, orgID, found)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]int, 0)

	for rows.Next() {
		var member int

		if err := rows.Scan(&rOrgID, &member); err != nil {
			return nil, err
		}

		if rOrgID != orgID {
			return nil, ErrMismatchOrgID
		}

		members = append(members, member)
	}

	return &am.Role{
		OrgID:    orgID,
		ID:       found,
		RoleName: foundName,
		Members:  members,
	}, nil
}
