package user

import (
	"context"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
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

// Get user by their email address
func (s *Service) Get(ctx context.Context, userContext am.UserContext, userEmail string) (oid int, user *am.User, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserSystem, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByEmail", userContext.GetOrgID(), userEmail))
}

// GetWithOrgID for internal lookups instead of getting orgID from context we pass it directly (system/support only).
func (s *Service) GetWithOrgID(ctx context.Context, userContext am.UserContext, orgID int, userEmail string) (oid int, user *am.User, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserSystem, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByEmail", orgID, userEmail))
}

// Get user by ID
func (s *Service) GetByID(ctx context.Context, userContext am.UserContext, userID int) (oid int, user *am.User, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserSystem, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByID", userContext.GetOrgID(), userID))
}

// GetByCID user by user custom id
func (s *Service) GetByCID(ctx context.Context, userContext am.UserContext, userCID string) (oid int, user *am.User, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByCID", userContext.GetOrgID(), userCID))
}

// get executes the scan against the previously created queryrow
func (s *Service) get(ctx context.Context, userContext am.UserContext, row *pgx.Row) (oid int, user *am.User, err error) {
	user = &am.User{}
	err = row.Scan(&user.OrgID, &user.UserID, &user.UserCID, &user.UserEmail, &user.FirstName, &user.LastName, &user.StatusID, &user.CreationTime, &user.Deleted)
	return user.OrgID, user, err
}

// List all users that match the supplied filter
func (s *Service) List(ctx context.Context, userContext am.UserContext, filter *am.UserFilter) (oid int, users []*am.User, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	var rows *pgx.Rows

	users = make([]*am.User, 0)
	oid = filter.OrgID

	if filter.WithDeleted {
		rows, err = s.pool.Query("userListWithDelete", oid, filter.DeletedValue, filter.Start, filter.Limit)
	} else {
		rows, err = s.pool.Query("userList", oid, filter.Start, filter.Limit)
	}

	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		user := &am.User{}

		if err := rows.Scan(&user.OrgID, &user.UserID, &user.UserCID, &user.UserEmail, &user.FirstName, &user.LastName, &user.StatusID, &user.CreationTime, &user.Deleted); err != nil {
			return 0, nil, err
		}

		if user.OrgID != oid {
			return 0, nil, am.ErrOrgIDMismatch
		}

		users = append(users, user)
	}

	return oid, users, nil
}

// Create a new user, set status to awaiting activation.
func (s *Service) Create(ctx context.Context, userContext am.UserContext, user *am.User) (oid int, uid int, ucid string, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "create") {
		return 0, 0, "", am.ErrUserNotAuthorized
	}
	oid = userContext.GetOrgID()

	tx, err := s.pool.Begin()
	defer tx.Rollback()

	err = tx.QueryRow("userExists", oid, -1, user.UserEmail).Scan(&oid, &uid, &ucid)
	if err != nil && err != pgx.ErrNoRows {
		return 0, 0, "", err
	}

	if uid != 0 {
		return 0, 0, "", am.ErrUserExists
	}

	id, err := uuid.NewV4()
	if err != nil {
		return 0, 0, "", err
	}
	ucid = id.String()
	now := time.Now().UnixNano()

	row := tx.QueryRow("userCreate", oid, ucid, user.UserEmail, user.FirstName, user.LastName, am.UserStatusAwaitActivation, now)
	if err := row.Scan(&oid, &uid, &ucid); err != nil {
		return 0, 0, "", err
	}

	return oid, uid, ucid, tx.Commit()
}

// Update allows the customer to update the details of their user
func (s *Service) Update(ctx context.Context, userContext am.UserContext, user *am.User, userID int) (oid int, uid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "update") {
		return 0, 0, am.ErrUserNotAuthorized
	}
	var tx *pgx.Tx

	tx, err = s.pool.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	oid, current, err := s.get(ctx, userContext, tx.QueryRow("userByID", userContext.GetOrgID(), userID))
	if err != nil {
		return 0, 0, err
	}

	fname := current.FirstName
	if user.FirstName != "" && user.FirstName != current.FirstName {
		fname = user.FirstName
	}

	lname := current.LastName
	if user.LastName != "" && user.LastName != current.LastName {
		lname = user.LastName
	}

	userStatusID := current.StatusID
	if user.StatusID != 0 && user.StatusID != current.StatusID {
		userStatusID = user.StatusID
	}
	row := tx.QueryRow("userUpdate", fname, lname, userStatusID, userContext.GetOrgID(), userID)
	if err := row.Scan(&oid, &uid); err != nil {
		return 0, 0, err
	}
	return oid, uid, tx.Commit()
}

// Delete the user
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, userID int) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "delete") {
		return 0, am.ErrUserNotAuthorized
	}

	err = s.pool.QueryRow("userDelete", userContext.GetOrgID(), userID).Scan(&oid)
	return oid, err
}
