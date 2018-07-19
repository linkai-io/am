package user

import (
	"context"
	"log"

	"github.com/jackc/pgx"
	uuid "github.com/satori/go.uuid"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/auth"
)

// Service for interfacing with postgresql/rds
type Service struct {
	pool       *pgx.ConnPool
	config     *pgx.ConnPoolConfig
	authorizer auth.Authorizer
}

// New returns an empty Service
func New(authorizer auth.Authorizer) *Service {
	return &Service{authorizer: authorizer}
}

// Init by parsing the config and initializing the database pool
func (s *Service) Init(config []byte) error {
	var err error

	s.config, err = s.parseConfig(config)
	if err != nil {
		return err
	}

	if s.pool, err = pgx.NewConnPool(*s.config); err != nil {
		return err
	}

	return nil
}

// parseConfig parses the configuration options and validates they are sane.
func (s *Service) parseConfig(config []byte) (*pgx.ConnPoolConfig, error) {
	dbstring := string(config)
	if dbstring == "" {
		return nil, am.ErrEmptyDBConfig
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		return nil, am.ErrInvalidDBString
	}

	return &pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 50,
		AfterConnect:   s.afterConnect,
	}, nil
}

// afterConnect will iterate over prepared statements with keywords
func (s *Service) afterConnect(conn *pgx.Conn) error {
	for k, v := range queryMap {
		if _, err := conn.Prepare(k, v); err != nil {
			log.Printf("%s\n", k)
			return err
		}
	}
	return nil
}

// IsAuthorized checks if an action is allowed by a particular user
func (s *Service) IsAuthorized(ctx context.Context, userContext am.UserContext, resource, action string) bool {
	if err := s.authorizer.IsUserAllowed(userContext.GetOrgID(), userContext.GetUserID(), resource, action); err != nil {
		return false
	}
	return true
}

// Get organization by organization name, system user only.
func (s *Service) Get(ctx context.Context, userContext am.UserContext, userID int) (*am.User, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserSystem, "read") {
		return nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByID", userContext.GetOrgID(), userID))
}

// GetByCID user by user custom id
func (s *Service) GetByCID(ctx context.Context, userContext am.UserContext, userCID string) (*am.User, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "read") {
		return nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByCID", userContext.GetOrgID(), userCID))
}

// get executes the scan against the previously created queryrow
func (s *Service) get(ctx context.Context, userContext am.UserContext, row *pgx.Row) (*am.User, error) {
	user := &am.User{}
	err := row.Scan(&user.OrgID, &user.UserID, &user.UserCID, &user.UserEmail, &user.FirstName, &user.LastName, &user.Deleted)
	return user, err
}

// List all users that match the supplied filter
func (s *Service) List(ctx context.Context, userContext am.UserContext, filter *am.UserFilter) ([]*am.User, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "read") {
		return nil, am.ErrUserNotAuthorized
	}
	users := make([]*am.User, 0)

	rows, err := s.pool.Query("userList", userContext.GetOrgID(), filter.Start, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		user := &am.User{}

		if err := rows.Scan(&user.OrgID, &user.UserID, &user.UserCID, &user.UserEmail, &user.FirstName, &user.LastName, &user.Deleted); err != nil {
			return nil, err
		}

		if user.OrgID != userContext.GetOrgID() {
			return nil, am.ErrOrgIDMismatch
		}

		users = append(users, user)
	}

	return users, nil
}

// Create a new user.
func (s *Service) Create(ctx context.Context, userContext am.UserContext, user *am.User) (userCID string, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "create") {
		return "", am.ErrUserNotAuthorized
	}
	tx, err := s.pool.Begin()
	defer tx.Rollback()

	oid := 0
	uid := 0
	cid := ""
	err = tx.QueryRow("userExists", user.OrgID, -1, user.UserEmail).Scan(&oid, &uid, &cid)
	if err != nil && err != pgx.ErrNoRows {
		return "", err
	}

	if uid != 0 {
		return "", am.ErrUserExists
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	userCID = id.String()

	if _, err = tx.Exec("userCreate", user.OrgID, userCID, user.UserEmail, user.FirstName, user.LastName); err != nil {

		return "", err
	}

	err = tx.Commit()
	return userCID, err
}

// Update allows the customer to update the details of their organization
func (s *Service) Update(ctx context.Context, userContext am.UserContext, user *am.User) error {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "update") {
		return am.ErrUserNotAuthorized
	}
	return nil
}

// Delete the organization
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, userID int) error {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "delete") {
		return am.ErrUserNotAuthorized
	}

	_, err := s.pool.Exec("userDelete", userID)
	return err
}
