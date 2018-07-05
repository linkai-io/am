package ladonrolemanager

import (
	"errors"

	"github.com/jackc/pgx"
	uuid "github.com/satori/go.uuid"
	"gopkg.linkai.io/v1/repos/am/am"
)

var (
	// ErrInvalidDriver returned if driver is not postgres or mysql
	ErrInvalidDriver = errors.New("invalid drivername specified, must be mysql or postgres, pg, pgx")
	// ErrTxCreateFailed returned if we could not begin a transaction
	ErrTxCreateFailed = errors.New("could not begin transaction")
	// ErrRoleNotFound unable to find role for orgID
	ErrRoleNotFound = errors.New("role id does not exist")
	// ErrMismatchOrgID if a query ever returns data that doesn't match the requester orgID
	ErrMismatchOrgID = errors.New("org id returned does not match specified org id")
)

// LadonRoleManager manages user roles and groups
type LadonRoleManager struct {
	db         *pgx.ConnPool
	driverName string
	stmts      *Statements
}

// New LadonRoleManager
func New(db *pgx.ConnPool, driverName string) *LadonRoleManager {
	return &LadonRoleManager{db: db, driverName: driverName}
}

// Init this role manager with a pgx connection pool
// TODO: test db tables exist
func (r *LadonRoleManager) Init() error {
	if r.stmts == nil {
		r.stmts = GetStatements(r.driverName)
		if r.stmts == nil {
			return ErrInvalidDriver
		}
	}
	return nil
}

// CreateRole with or without initial members returns roleID
func (r *LadonRoleManager) CreateRole(g *am.Role) (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	g.ID = id.String()

	if _, err = r.db.Exec(r.stmts.QueryInsertRole, g.OrgID, g.ID); err != nil {
		return "", err
	}
	return g.ID, r.AddMembers(g.OrgID, g.ID, g.Members)
}

// AddMembers iterates over every member for the group/roleid and adds it to the ladon_role_members table
func (r *LadonRoleManager) AddMembers(orgID int32, roleID string, members []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for _, member := range members {
		if _, err := tx.Exec(r.stmts.QueryAddMembers, orgID, roleID, member); err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return nil
}

// RemoveMembers iterates over every member for the group/roleid and removes it from the ladon_role_members table
func (r *LadonRoleManager) RemoveMembers(orgID int32, roleID string, members []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for _, member := range members {
		if _, err := tx.Exec(r.stmts.QueryDeleteMembers, orgID, roleID, member); err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return nil
}

// FindByMember returns roles that a member belongs to
func (r *LadonRoleManager) FindByMember(orgID int32, member string, limit, offset int) ([]*am.Role, error) {
	var roleIDs []string
	rows, err := r.db.Query(r.stmts.QueryFindByMember, orgID, member, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var rOrgID int32

		if err := rows.Scan(&rOrgID, &id); err != nil {
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

// Get a role specified by roleID for the orgID
func (r *LadonRoleManager) Get(orgID int32, roleID string) (*am.Role, error) {
	var found string
	var rOrgID int32

	if err := r.db.QueryRow(r.stmts.QueryGetRole, orgID, roleID).Scan(&rOrgID, &found); err != nil {
		return nil, err
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

	members := make([]string, 0)

	for rows.Next() {
		var member string

		if err := rows.Scan(&rOrgID, &member); err != nil {
			return nil, err
		}

		if rOrgID != orgID {
			return nil, ErrMismatchOrgID
		}

		members = append(members, member)
	}

	return &am.Role{
		OrgID:   orgID,
		ID:      found,
		Members: members,
	}, nil
}

// List all roles for an organization
func (r *LadonRoleManager) List(orgID int32, limit, offset int) ([]*am.Role, error) {
	var roleIDs []string
	rows, err := r.db.Query(r.stmts.QueryList, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var rOrgID int32

		if err := rows.Scan(&rOrgID, &id); err != nil {
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
