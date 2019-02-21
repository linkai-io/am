package user

import (
	"context"
	"time"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
	"github.com/rs/zerolog/log"
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
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "userservice.Get").
		Str("TraceID", userContext.GetTraceID()).Logger()
	serviceLog.Info().Str("userEmail_parameter", userEmail).Msg("processing")

	if !s.IsAuthorized(ctx, userContext, am.RNUserSystem, "read") {
		serviceLog.Error().Msg("user not authorized")
		return 0, nil, am.ErrUserNotAuthorized
	}

	return s.get(ctx, userContext, s.pool.QueryRow("userByEmail", userContext.GetOrgID(), userEmail))
}

// GetWithOrgID for internal lookups instead of getting orgID from context we pass it directly (system/support only).
func (s *Service) GetWithOrgID(ctx context.Context, userContext am.UserContext, orgID int, userCID string) (oid int, user *am.User, err error) {
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "userservice.GetWithOrgID").
		Str("TraceID", userContext.GetTraceID()).Logger()
	serviceLog.Info().Int("orgID_parameter", orgID).Str("userCID_parameter", userCID).Msg("processing")

	if !s.IsAuthorized(ctx, userContext, am.RNUserSystem, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByCID", orgID, userCID))
}

// GetByID to be called with system context
func (s *Service) GetByID(ctx context.Context, userContext am.UserContext, userID int) (oid int, user *am.User, err error) {
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "userservice.GetByID").
		Str("TraceID", userContext.GetTraceID()).Logger()
	serviceLog.Info().Int("userID_parameter", userID).Msg("processing")

	if !s.IsAuthorized(ctx, userContext, am.RNUserSystem, "read") {
		serviceLog.Error().Msg("user not authorized")
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByID", userContext.GetOrgID(), userID))
}

// GetByCID user by user custom id
func (s *Service) GetByCID(ctx context.Context, userContext am.UserContext, userCID string) (oid int, user *am.User, err error) {
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "userservice.GetByCID").
		Str("TraceID", userContext.GetTraceID()).Logger()
	serviceLog.Info().Str("userCID_parameter", userCID).Msg("processing")

	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "read") {
		serviceLog.Error().Msg("user not authorized")
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("userByCID", userContext.GetOrgID(), userCID))
}

// get executes the scan against the previously created queryrow
func (s *Service) get(ctx context.Context, userContext am.UserContext, row *pgx.Row) (oid int, user *am.User, err error) {
	user = &am.User{}
	var createTime time.Time
	err = row.Scan(&user.OrgID, &user.UserID, &user.UserCID, &user.UserEmail, &user.FirstName, &user.LastName, &user.StatusID, &createTime, &user.Deleted)
	if err == pgx.ErrNoRows {
		return 0, nil, am.ErrNoResults
	}
	user.CreationTime = createTime.UnixNano()
	return user.OrgID, user, err
}

// List all users that match the supplied filter. If filter.OrgID is different than user context, ensure system user is calling.
func (s *Service) List(ctx context.Context, userContext am.UserContext, filter *am.UserFilter) (oid int, users []*am.User, err error) {
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "userservice.List").
		Str("TraceID", userContext.GetTraceID()).Logger()
	serviceLog.Info().Int("filterOrgID_parameter", filter.OrgID).Int("filterStart_parameter", filter.Start).Int("filterLimit_parameter", filter.Limit).Msg("processing")

	if filter.OrgID != userContext.GetOrgID() {
		if !s.IsAuthorized(ctx, userContext, am.RNUserSystem, "read") {
			serviceLog.Error().Msg("user not authorized orgID is not context orgid")
			return 0, nil, am.ErrUserNotAuthorized
		}
		oid = filter.OrgID
	} else {
		if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "read") {
			serviceLog.Error().Msg("user not authorized")
			return 0, nil, am.ErrUserNotAuthorized
		}
		oid = userContext.GetOrgID()
	}

	var rows *pgx.Rows

	users = make([]*am.User, 0)

	if filter.WithDeleted {
		rows, err = s.pool.Query("userListWithDelete", oid, filter.DeletedValue, filter.Start, filter.Limit)
	} else {
		rows, err = s.pool.Query("userList", oid, filter.Start, filter.Limit)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil, am.ErrNoResults
		}
		return 0, nil, err
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		var createTime time.Time
		user := &am.User{}

		if err := rows.Scan(&user.OrgID, &user.UserID, &user.UserCID, &user.UserEmail, &user.FirstName, &user.LastName, &user.StatusID, &createTime, &user.Deleted); err != nil {
			return 0, nil, err
		}

		if user.OrgID != oid {
			return 0, nil, am.ErrOrgIDMismatch
		}
		user.CreationTime = createTime.UnixNano()
		users = append(users, user)
	}

	return oid, users, nil
}

// Create a new user, set status to active
func (s *Service) Create(ctx context.Context, userContext am.UserContext, user *am.User) (oid int, uid int, ucid string, err error) {
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "userservice.Create").
		Str("TraceID", userContext.GetTraceID()).Logger()
	serviceLog.Info().Str("userEmail_parameter", user.UserEmail).Str("userCID_parameter", user.UserCID).Msg("processing")

	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "create") {
		serviceLog.Error().Msg("user not authorized")
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

	if user.UserCID == "" {
		return 0, 0, "", am.ErrUserCIDEmpty
	}
	now := time.Now()

	row := tx.QueryRow("userCreate", oid, user.UserCID, user.UserEmail, user.FirstName, user.LastName, am.UserStatusActive, now)
	if err := row.Scan(&oid, &uid, &ucid); err != nil {
		return 0, 0, "", err
	}

	return oid, uid, ucid, tx.Commit()
}

// Update allows the customer to update the details of their user. If userID does not equal user context, ensure they have UserManage
// permissions.
func (s *Service) Update(ctx context.Context, userContext am.UserContext, user *am.User, userID int) (oid int, uid int, err error) {
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "userservice.Update").
		Str("TraceID", userContext.GetTraceID()).Logger()
	serviceLog.Info().Int("userID_parameter", userID).Msg("processing")

	if userContext.GetUserID() != userID {
		if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "update") {
			serviceLog.Error().Msg("user not authorized userid does not equal context userid")
			return 0, 0, am.ErrUserNotAuthorized
		}
	} else {
		if !s.IsAuthorized(ctx, userContext, am.RNUserSelf, "update") {
			serviceLog.Error().Msg("user not authorized")
			return 0, 0, am.ErrUserNotAuthorized
		}
	}

	var tx *pgx.Tx

	tx, err = s.pool.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	oid, current, err := s.get(ctx, userContext, tx.QueryRow("userByID", userContext.GetOrgID(), userID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, am.ErrNoResults
		}
		return 0, 0, err
	}

	// only system users can change UserCID
	userCID := current.UserCID
	if user.UserCID != "" && user.UserCID != current.UserCID {
		if !s.IsAuthorized(ctx, userContext, am.RNSystem, "update") {
			return 0, 0, am.ErrUserNotAuthorized
		}
		userCID = user.UserCID
	}

	email := current.UserEmail
	if user.UserEmail != "" && user.UserEmail != current.UserEmail {
		email = user.UserEmail
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

	row := tx.QueryRow("userUpdate", userCID, email, fname, lname, userStatusID, userContext.GetOrgID(), userID)
	if err := row.Scan(&oid, &uid); err != nil {
		return 0, 0, err
	}
	return oid, uid, tx.Commit()
}

// Delete the user
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, userID int) (oid int, err error) {
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "userservice.Delete").
		Str("TraceID", userContext.GetTraceID()).Logger()
	serviceLog.Info().Int("userID_parameter", userID).Msg("processing")

	if !s.IsAuthorized(ctx, userContext, am.RNUserManage, "delete") {
		serviceLog.Error().Msg("user not authorized")
		return 0, am.ErrUserNotAuthorized
	}

	err = s.pool.QueryRow("userDelete", userContext.GetOrgID(), userID).Scan(&oid)
	return oid, err
}
