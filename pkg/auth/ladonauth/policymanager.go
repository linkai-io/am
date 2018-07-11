package ladonauth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/ory/ladon"
	"github.com/ory/ladon/compiler"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidDriver returned if driver is not postgres or mysql
	ErrInvalidDriver = errors.New("invalid drivername specified, must be mysql or postgres, pg, pgx")
)

// LadonPolicyManager implements the ladon/Manager without requiring sqlx or migrations packages
type LadonPolicyManager struct {
	db         *pgx.ConnPool
	driverName string
	stmts      *PolicyStatements
}

// NewPolicyManager creates a new, uninitialized LadonPolicyManager
func NewPolicyManager(db *pgx.ConnPool, driverName string) *LadonPolicyManager {
	return &LadonPolicyManager{db: db, driverName: driverName}
}

// SetStatements allows callers to just provide their own statements if they
// want to support something other than postgres/mysql
// Note you must call this before Init() if you wish to override the driver specific
// statements.
func (s *LadonPolicyManager) SetStatements(statements *PolicyStatements) {
	s.stmts = statements
}

// Init ensures statements are properly mapped
func (s *LadonPolicyManager) Init() error {
	if s.stmts == nil {
		s.stmts = GetPolicyStatements(s.driverName)
		if s.stmts == nil {
			return ErrInvalidDriver
		}
	}
	return nil
}

// Update updates a policy in the database by deleting original and re-creating
func (s *LadonPolicyManager) Update(policy ladon.Policy) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.delete(policy.GetID(), tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err := s.create(policy, tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	return nil
}

// Create inserts a new policy
func (s *LadonPolicyManager) Create(policy ladon.Policy) (err error) {

	tx, err := s.db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.create(policy, tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	return nil
}

func (s *LadonPolicyManager) create(policy ladon.Policy, tx *pgx.Tx) (err error) {
	conditions := []byte("{}")
	if policy.GetConditions() != nil {
		cs := policy.GetConditions()
		conditions, err = json.Marshal(&cs)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	meta := []byte("{}")
	if policy.GetMeta() != nil {
		meta = policy.GetMeta()
	}

	if _, err = tx.Exec(s.stmts.QueryInsertPolicy, policy.GetID(), policy.GetDescription(), policy.GetEffect(), conditions, meta); err != nil {
		return errors.WithStack(err)
	}

	type relation struct {
		p []string
		t string
	}
	var relations = []relation{
		{p: policy.GetActions(), t: "action"},
		{p: policy.GetResources(), t: "resource"},
		{p: policy.GetSubjects(), t: "subject"},
	}

	for _, rel := range relations {
		var query string
		var queryRel string

		switch rel.t {
		case "action":
			query = s.stmts.QueryInsertPolicyActions
			queryRel = s.stmts.QueryInsertPolicyActionsRel
		case "resource":
			query = s.stmts.QueryInsertPolicyResources
			queryRel = s.stmts.QueryInsertPolicyResourcesRel
		case "subject":
			query = s.stmts.QueryInsertPolicySubjects
			queryRel = s.stmts.QueryInsertPolicySubjectsRel
		}

		for _, template := range rel.p {
			h := sha256.New()
			h.Write([]byte(template))
			id := fmt.Sprintf("%x", h.Sum(nil))

			compiled, err := compiler.CompileRegex(template, policy.GetStartDelimiter(), policy.GetEndDelimiter())
			if err != nil {
				return errors.WithStack(err)
			}

			if _, err := tx.Exec(query, id, template, compiled.String(), strings.Index(template, string(policy.GetStartDelimiter())) >= -1); err != nil {
				return errors.WithStack(err)
			}
			if _, err := tx.Exec(queryRel, policy.GetID(), id); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

// FindRequestCandidates returns policies that potentially match a ladon.Request
func (s *LadonPolicyManager) FindRequestCandidates(r *ladon.Request) (ladon.Policies, error) {
	rows, err := s.db.Query(s.stmts.QueryRequestCandidates, r.Subject, r.Subject)
	if err == sql.ErrNoRows {
		return nil, ladon.NewErrResourceNotFound(err)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func scanRows(rows *pgx.Rows) (ladon.Policies, error) {
	var policies = map[string]*ladon.DefaultPolicy{}

	for rows.Next() {
		var p ladon.DefaultPolicy
		var conditions []byte
		var resource, subject, action sql.NullString
		p.Actions = []string{}
		p.Subjects = []string{}
		p.Resources = []string{}

		if err := rows.Scan(&p.ID, &p.Effect, &conditions, &p.Description, &p.Meta, &subject, &resource, &action); err == sql.ErrNoRows {
			return nil, ladon.NewErrResourceNotFound(err)
		} else if err != nil {
			return nil, errors.WithStack(err)
		}

		p.Conditions = ladon.Conditions{}
		if err := json.Unmarshal(conditions, &p.Conditions); err != nil {
			return nil, errors.WithStack(err)
		}

		if c, ok := policies[p.ID]; ok {
			if action.Valid {
				policies[p.ID].Actions = append(c.Actions, action.String)
			}

			if subject.Valid {
				policies[p.ID].Subjects = append(c.Subjects, subject.String)
			}

			if resource.Valid {
				policies[p.ID].Resources = append(c.Resources, resource.String)
			}
		} else {
			if action.Valid {
				p.Actions = []string{action.String}
			}

			if subject.Valid {
				p.Subjects = []string{subject.String}
			}

			if resource.Valid {
				p.Resources = []string{resource.String}
			}

			policies[p.ID] = &p
		}
	}

	var result = make(ladon.Policies, len(policies))
	var count int
	for _, v := range policies {
		v.Actions = uniq(v.Actions)
		v.Resources = uniq(v.Resources)
		v.Subjects = uniq(v.Subjects)
		result[count] = v
		count++
	}

	return result, nil
}

// GetAll returns all policies
func (s *LadonPolicyManager) GetAll(limit, offset int64) (ladon.Policies, error) {
	rows, err := s.db.Query(s.stmts.GetAllQuery, limit, offset)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// Get retrieves a policy.
func (s *LadonPolicyManager) Get(id string) (ladon.Policy, error) {
	rows, err := s.db.Query(s.stmts.GetQuery, id)
	if err == sql.ErrNoRows {
		return nil, ladon.NewErrResourceNotFound(err)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	policies, err := scanRows(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if len(policies) == 0 {
		return nil, ladon.NewErrResourceNotFound(sql.ErrNoRows)
	}

	return policies[0], nil
}

// Delete removes a policy.
func (s *LadonPolicyManager) Delete(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.delete(id, tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	return nil
}

// Delete removes a policy.
func (s *LadonPolicyManager) delete(id string, tx *pgx.Tx) error {
	_, err := tx.Exec(s.stmts.DeletePolicy, id)
	return errors.WithStack(err)
}

func uniq(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}
